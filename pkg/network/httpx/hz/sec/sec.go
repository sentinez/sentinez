// Copyright 2025 Sentinez Labs.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpxhzsec

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/experimental"
	"github.com/corazawaf/coraza/v3/types"
	httpxhz "github.com/sentinez/sentinez/pkg/network/httpx/hz"
	"github.com/sentinez/sentinez/pkg/x/errorx"
	"github.com/sentinez/sentinez/pkg/zlog"
)

func decorNewTransaction(
	waf coraza.WAF) func(*httpxhz.Context) types.Transaction {

	newTX := func(*httpxhz.Context) types.Transaction {
		return waf.NewTransaction()
	}

	if ctxWAF, ok := waf.(experimental.WAFWithOptions); ok {
		newTX = func(ctx *httpxhz.Context) types.Transaction {
			return ctxWAF.NewTransactionWithOptions(experimental.Options{
				Context: ctx.Context(),
			})
		}
	}

	return newTX
}

// nolint:funlen
func WrapHandlerWithCallback(waf coraza.WAF, next httpxhz.RequestHandler,
	cb func(*httpxhz.Context, types.Transaction)) httpxhz.RequestHandler {
	if waf == nil {
		return next
	}
	newTX := decorNewTransaction(waf)

	return func(ctx *httpxhz.Context) error {
		tx := newTX(ctx)
		defer postProcess(ctx, tx, cb)

		if tx.IsRuleEngineOff() {
			return next(ctx)
		}

		if err := processRequestHandler(ctx, tx); err != nil {
			debugLogger(tx, err, "failed to process request")
			return nil
		}

		err := next(ctx)

		if err := processResponseHandler(ctx, tx); err != nil {
			debugLogger(tx, err, "failed to process response")
			return nil
		}

		return err
	}
}

func postProcess(ctx *httpxhz.Context, tx types.Transaction,
	callback func(*httpxhz.Context, types.Transaction)) {
	// final phase
	tx.ProcessLogging()

	if callback != nil {
		callback(ctx, tx)
	}

	if err := tx.Close(); err != nil {
		debugLogger(tx, err, "failed to close transaction")
	}
}

func processRequestHandler(ctx *httpxhz.Context, tx types.Transaction) error {
	if it, err := processRequest(ctx, tx); err != nil {
		zlog.Debugf("failed to process request: %v", err)
		return err
	} else if it != nil {

		code := obtainStatusCodeFromInterruptionOrDefault(it,
			ctx.Response.StatusCode(),
		)

		ctx.SetStatusCode(code)
		if code == http.StatusForbidden {
			_ = httpxhz.Forbidden(ctx)
		}

		return errorx.F("[interrupted][request] with code: %d", code)
	}

	return nil
}

func debugLogger(tx types.Transaction, err error, msg string) {
	tx.DebugLogger().
		Error().
		Err(err).
		Msg(msg)
}

// processRequest ...
// ref: https://github.com/corazawaf/coraza/blob/main/http/middleware.go#L27
func processRequest(ctx *httpxhz.Context,
	tx types.Transaction) (*types.Interruption, error) {

	if it := processRequestHeader(ctx, tx); it != nil {
		return it, nil
	}

	if it, err := processRequestBody(ctx, tx); err != nil {
		return nil, err

	} else if it != nil {
		return it, nil
	}

	return nil, nil
}

func processRequestHeader(ctx *httpxhz.Context,
	tx types.Transaction) *types.Interruption {

	processRequestConnection(ctx, tx)

	host := string(ctx.Host())
	if host != "" {
		tx.AddRequestHeader("Host", host)
		tx.SetServerName(host)
	}

	transferEncoding := ctx.Request.Header.Get("Transfer-Encoding")
	if transferEncoding != "" {
		tx.AddRequestHeader("Transfer-Encoding", transferEncoding)
	}

	in := tx.ProcessRequestHeaders()
	if in != nil {
		return in
	}

	return nil
}

func processRequestConnection(ctx *httpxhz.Context, tx types.Transaction) {

	var client string
	var cport int

	idx := strings.LastIndexByte(ctx.RemoteAddr().String(), ':')
	if idx != -1 {
		client = ctx.RemoteAddr().String()[:idx]
		cport, _ = strconv.Atoi(ctx.RemoteAddr().String()[idx+1:])
	}

	tx.ProcessConnection(client, cport, "", 0)
	tx.ProcessURI(
		ctx.URI().String(),
		string(ctx.Method()),
		ctx.Request.Header.GetProtocol(),
	)
	ctx.Request.Header.VisitAll(func(k, v []byte) {
		tx.AddRequestHeader(string(k), string(v))
	})
}

func processRequestBody(ctx *httpxhz.Context,
	tx types.Transaction) (*types.Interruption, error) {

	if tx.IsRequestBodyAccessible() {
		if it, err := canRequestBodyAccessible(ctx, tx); err != nil {
			return nil, err
		} else if it != nil {
			return it, nil
		}
	}

	return tx.ProcessRequestBody()
}

func canRequestBodyAccessible(ctx *httpxhz.Context,
	tx types.Transaction) (*types.Interruption, error) {

	body, err := ctx.Body()
	if err != nil {
		return nil, fmt.Errorf("failed to get body: %v", err)
	}

	if len(ctx.Request.Body()) != 0 {
		it, _, err := tx.ReadRequestBodyFrom(ctx.RequestBodyStream())
		if err != nil {
			return nil, fmt.Errorf("failed to append request body: %v", err)
		}

		if it != nil {
			return it, nil
		}

		ctx.Request.SetBody(body)
	}

	return nil, nil
}

func processResponseHandler(ctx *httpxhz.Context, tx types.Transaction) error {
	if tx.IsInterrupted() {
		return nil
	}

	i := interceptor{tx: tx, proto: ctx.Request.Header.GetProtocol()}
	it, err := i.WriteResponseBody(ctx)
	if err != nil {
		return err
	}

	if it != nil {
		ctx.Response.Reset()
		code := obtainStatusCodeFromInterruptionOrDefault(it,
			ctx.Response.StatusCode(),
		)

		ctx.Response.SetStatusCode(code)
		if code == http.StatusForbidden {
			_ = httpxhz.Forbidden(ctx)
		}

		return errorx.F("[interrupted][response] with code: %d", code)
	}
	return releaseBodyReader(ctx, tx)
}

func releaseBodyReader(ctx *httpxhz.Context, tx types.Transaction) error {

	reader, err := tx.ResponseBodyReader()
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return fmt.Errorf("failed to release resp body reader: %v", err)
	}

	if _, err = io.Copy(ctx, reader); err != nil {
		return fmt.Errorf("failed to copy the resp body: %v", err)
	}

	return nil
}
