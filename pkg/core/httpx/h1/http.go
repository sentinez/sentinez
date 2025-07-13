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

package httpx1

import (
	"io"
	"net/http"

	"github.com/sentinez/sentinez/pkg/std/zlog"
)

func HandlerFunc(path string, handler func(Context) error) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := Context{
			req:  r,
			resp: w,
		}

		if err := handler(ctx); err != nil {
			zlog.Errorf(
				"httpx1: error in path %s: %v", path, err)
		}
	})
}

func ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, nil)
}

func Shutdown() error {
	// Note: http.Server does not have a Shutdown method in the standard library
	// This is a placeholder for future implementation if needed.
	return nil
}

// Do acts as a proxy: forwards the incoming request to the target URI and
// returns the response.
func Do(ctx Context, uri string) error {
	req, err := http.NewRequestWithContext(
		ctx.req.Context(), ctx.req.Method, uri, ctx.req.Body)
	if err != nil {
		zlog.Errorf("httpx1: failed to create request: %v", err)
		return err
	}
	for name, values := range ctx.req.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zlog.Errorf("httpx1: failed to perform request: %v", err)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	ctx.resp.WriteHeader(resp.StatusCode)
	for name, values := range resp.Header {
		for _, value := range values {
			ctx.resp.Header().Add(name, value)
		}
	}
	if _, err := io.Copy(ctx.resp, resp.Body); err != nil {
		zlog.Errorf("httpx1: failed to copy response body: %v", err)
		return err
	}
	return nil
}

func Convert(handler func(ctx Context) error,
) func(resp http.ResponseWriter, req *http.Request) {

	return func(resp http.ResponseWriter, req *http.Request) {
		ctx := Context{resp: resp, req: req}
		if err := handler(ctx); err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
		}
	}
}
