// Copyright 2025 Sentinez Labs.
// Copyright 2022 Juan Pablo Tosso and the OWASP Coraza contributors
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

package http2sec

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/experimental"
	"github.com/corazawaf/coraza/v3/types"
	"github.com/sentinez/sentinez/pkg/std/zlog"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func decorNewTransaction(waf coraza.WAF) func(*http.Request) types.Transaction {

	newTX := func(*http.Request) types.Transaction {
		return waf.NewTransaction()
	}

	if ctxwaf, ok := waf.(experimental.WAFWithOptions); ok {
		newTX = func(r *http.Request) types.Transaction {
			return ctxwaf.NewTransactionWithOptions(experimental.Options{
				Context: r.Context(),
			})
		}
	}

	return newTX
}

func convertRequestContext(ctx *fasthttp.RequestCtx, r *http.Request) {
	if err := fasthttpadaptor.ConvertRequest(ctx, r, true); err != nil {
		ctx.Error("failed to convert request context",
			fasthttp.StatusInternalServerError)
		return
	}
}

func WrapHandler(
	waf coraza.WAF, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	if waf == nil {
		return next
	}
	newTX := decorNewTransaction(waf)

	return func(ctx *fasthttp.RequestCtx) {
		r := new(http.Request)
		convertRequestContext(ctx, r)
		tx := newTX(r)
		defer postProcess(tx)

		if tx.IsRuleEngineOff() {
			next(ctx)
			return
		}

		processRequests := processRequestHandler(r)
		if err := processRequests(ctx, tx); err != nil {
			debugLogger(tx, err, "failed to process request")
			return
		}

		next(ctx)

		if err := processResponse(ctx, tx); err != nil {
			debugLogger(tx, err, "failed to process response")
			return
		}
	}
}

func postProcess(tx types.Transaction) {
	tx.ProcessLogging()
	if err := tx.Close(); err != nil {
		debugLogger(tx, err, "failed to close transaction")
	}
}

func processRequestHandler(r *http.Request,
) func(*fasthttp.RequestCtx, types.Transaction) error {

	return func(ctx *fasthttp.RequestCtx, tx types.Transaction) error {
		if it, err := processRequest(tx, r); err != nil {
			zlog.Debugf("failed to process request: %v", err)

			return err
		} else if it != nil {
			zlog.Debugf("[processing] req: action=%s, status=%d, ruleID=%d",
				it.Action, it.Status, it.RuleID)

			code := obtainStatusCodeFromInterruptionOrDefault(
				it,
				ctx.Response.StatusCode(),
			)
			zlog.Debugf("[block] %s: %d", ctx.Request.URI().RequestURI(), code)

			ctx.SetStatusCode(code)
			if _, err := ctx.Write([]byte("access denied")); err != nil {
				zlog.Errorf("failed to write response body: %v", err)
			}

			return fmt.Errorf("[interrupted] request with code: %d", code)
		}

		return nil
	}
}

func debugLogger(tx types.Transaction, err error, msg string) {
	tx.DebugLogger().
		Error().
		Err(err).
		Msg(msg)
}

// processRequest ...
// ref: https://github.com/corazawaf/coraza/blob/main/http/middleware.go#L27
func processRequest(tx types.Transaction,
	req *http.Request) (*types.Interruption, error) {

	if it := processRequestHeader(req, tx); it != nil {
		zlog.Debugf("security req header: action=%s, id=%d",
			it.Action, it.RuleID)
		return it, nil
	}

	if it, err := processRequestBody(req, tx); err != nil {
		return nil, err

	} else if it != nil {
		zlog.Debugf("security req body: action=%s, id=%d", it.Action, it.RuleID)
		return it, nil
	}

	return nil, nil
}

func processRequestHeader(req *http.Request,
	tx types.Transaction) *types.Interruption {

	processRequestConnection(req, tx)

	if req.Host != "" {
		tx.AddRequestHeader("Host", req.Host)
		tx.SetServerName(req.Host)
	}
	if req.TransferEncoding != nil {
		tx.AddRequestHeader("Transfer-Encoding", req.TransferEncoding[0])
	}

	in := tx.ProcessRequestHeaders()
	if in != nil {
		return in
	}

	return nil
}

func processRequestConnection(req *http.Request, tx types.Transaction) {

	var client string
	var cport int

	idx := strings.LastIndexByte(req.RemoteAddr, ':')
	if idx != -1 {
		client = req.RemoteAddr[:idx]
		cport, _ = strconv.Atoi(req.RemoteAddr[idx+1:])
	}

	tx.ProcessConnection(client, cport, "", 0)
	tx.ProcessURI(req.URL.String(), req.Method, req.Proto)
	for k, vr := range req.Header {
		for _, v := range vr {
			tx.AddRequestHeader(k, v)
		}
	}
}

func processRequestBody(req *http.Request,
	tx types.Transaction) (*types.Interruption, error) {

	if tx.IsRequestBodyAccessible() {
		if it, err := canRequestBodyAccessible(req, tx); err != nil {
			return nil, err
		} else if it != nil {
			return it, nil
		}
	}

	return tx.ProcessRequestBody()
}

func canRequestBodyAccessible(req *http.Request,
	tx types.Transaction) (*types.Interruption, error) {

	if req.Body != nil && req.Body != http.NoBody {
		it, _, err := tx.ReadRequestBodyFrom(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to append request body: %v", err)
		}

		if it != nil {
			return it, nil
		}

		rbr, err := tx.RequestBodyReader()
		if err != nil {
			return nil, fmt.Errorf("failed to get the request body: %v", err)
		}

		bodyReader := io.MultiReader(rbr, req.Body)
		req.Body = io.NopCloser(bodyReader)
	}

	return nil, nil
}

func processResponse(ctx *fasthttp.RequestCtx,
	tx types.Transaction) error {
	i := interceptor{tx: tx, proto: string(ctx.Request.Header.Protocol())}

	if tx.IsInterrupted() {
		return nil
	}

	it, err := i.WriteResponseBody(ctx)
	if err != nil {
		return err
	}

	if it != nil {
		ctx.Response.Reset()
		code := obtainStatusCodeFromInterruptionOrDefault(
			it,
			ctx.Response.StatusCode(),
		)
		ctx.Response.Header.Set("Content-Length", "0")
		ctx.Response.SetStatusCode(code)

		if _, err := ctx.Write([]byte("access denied")); err != nil {
			zlog.Errorf("failed to write response body: %v", err)
		}

		return nil
	}

	return releaseBodyReader(ctx, tx)
}

func releaseBodyReader(ctx *fasthttp.RequestCtx, tx types.Transaction) error {

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
