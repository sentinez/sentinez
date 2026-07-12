// Copyright 2025 Duc-Hung Ho.
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

package corers

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/experimental"
	"github.com/corazawaf/coraza/v3/types"
	"github.com/sentinez/core/common/bytestr"
	corehttp "github.com/sentinez/core/http"
	httpconst "github.com/sentinez/core/http/const"
	"github.com/sentinez/shared/bytesconv"
)

func processRequestHandler(ctx corehttp.Context, tx types.Transaction) error {
	if it, err := processRequest(ctx, tx); err != nil {
		return err
	} else if it != nil {

		code := obtainStatusCodeFromInterruptionOrDefault(it, ctx.StatusCode())

		ctx.SetStatusCode(code)

		return fmt.Errorf("[interrupted][request] with code: %d", code)
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
func processRequest(ctx corehttp.Context,
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

func processRequestHeader(ctx corehttp.Context,
	tx types.Transaction) *types.Interruption {

	processRequestConnection(ctx, tx)

	host := bytesconv.B2s(ctx.Host())
	if host != "" {
		tx.AddRequestHeader("Host", host)
		tx.SetServerName(host)
	}

	transferEncoding := ctx.Header(bytestr.HeaderTransferEncoding)
	if transferEncoding != nil {
		tx.AddRequestHeader(
			httpconst.HeaderTransferEncoding,
			bytesconv.B2s(transferEncoding),
		)
	}

	in := tx.ProcessRequestHeaders()
	if in != nil {
		return in
	}

	return nil
}

func processRequestConnection(ctx corehttp.Context, tx types.Transaction) {

	var client string
	var cport int

	remoteAddr := ctx.RemoteAddr()
	idx := bytes.LastIndexByte(remoteAddr, ':')
	if idx != -1 {
		client = bytesconv.B2s(remoteAddr[:idx])
		cport, _ = strconv.Atoi(bytesconv.B2s(remoteAddr[idx+1:]))
	} else {
		client = bytesconv.B2s(remoteAddr)
	}

	tx.ProcessConnection(client, cport, "", 0)
	tx.ProcessURI(
		bytesconv.B2s(ctx.URI()),
		bytesconv.B2s(ctx.Method()),
		ctx.Protocol(),
	)
	ctx.VisitRequestHeaders(func(k, v []byte) {
		tx.AddRequestHeader(bytesconv.B2s(k), bytesconv.B2s(v))
	})
}

func processRequestBody(ctx corehttp.Context,
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

func canRequestBodyAccessible(ctx corehttp.Context,
	tx types.Transaction) (*types.Interruption, error) {

	body := ctx.Body()

	if len(body) != 0 {
		it, _, err := tx.ReadRequestBodyFrom(ctx.RequestBodyStream())
		if err != nil {
			return nil, fmt.Errorf("failed to append request body: %v", err)
		}

		if it != nil {
			return it, nil
		}

		ctx.SetBody(body)
	}

	return nil, nil
}

func processResponseHandler(ctx corehttp.Context, tx types.Transaction) error {
	if tx.IsInterrupted() {
		return nil
	}

	i := interceptor{tx: tx, proto: ctx.Protocol()}
	it, err := i.writeResponseBody(ctx)
	if err != nil {
		return err
	}

	if it != nil {
		ctx.ResetResponse()
		code := obtainStatusCodeFromInterruptionOrDefault(it, ctx.StatusCode())

		ctx.SetStatusCode(code)

		return fmt.Errorf("[interrupted][response] with code: %d", code)
	}

	return releaseBodyReader(ctx, tx)
}

func releaseBodyReader(ctx corehttp.Context, tx types.Transaction) error {

	reader, err := tx.ResponseBodyReader()
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return fmt.Errorf("failed to release resp body reader: %v", err)
	}

	if err = ctx.Copy(reader); err != nil {
		return fmt.Errorf("failed to copy the resp body: %v", err)
	}

	return nil
}

func newTransaction(waf coraza.WAF, ctx corehttp.Context) types.Transaction {

	newTX := func(corehttp.Context) types.Transaction {
		return waf.NewTransaction()
	}

	if ctxWAF, ok := waf.(experimental.WAFWithOptions); ok {
		newTX = func(ctx corehttp.Context) types.Transaction {
			return ctxWAF.NewTransactionWithOptions(experimental.Options{
				Context: ctx.Context(),
			})
		}
	}

	return newTX(ctx)
}
