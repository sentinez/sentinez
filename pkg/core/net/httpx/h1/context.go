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
	"encoding/json"
	"net/http"

	"github.com/sentinez/sentinez/pkg/core/net/httpx"

	"github.com/gorilla/websocket"
)

var _ IContext = (*Context)(nil)

type IContext interface {
	httpx.Context
	Upgrade() (*websocket.Conn, error)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		_ = r       // Ignore the request for origin check
		return true // Allow all origins for simplicity
	},
}

type Context struct {
	req  *http.Request
	resp http.ResponseWriter
}

func (c *Context) Method() string {
	return c.req.Method
}

func (c *Context) Path() string {
	return c.req.URL.Path
}

func (c *Context) JSON(statusCode int, body []byte) error {
	c.resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.resp.WriteHeader(statusCode)

	// Use a JSON encoder to write the data
	encoder := json.NewEncoder(c.resp)
	return encoder.Encode(body)
}

func (c *Context) AsCore() *Context {
	return &Context{
		req:  c.req,
		resp: c.resp,
	}
}

func (c *Context) Request() *http.Request {
	return c.req
}

func (c *Context) Response() http.ResponseWriter {
	return c.resp
}

func (c *Context) Upgrade() (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(c.resp, c.req, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *Context) String(statusCode int, msg string) error {
	c.resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.resp.WriteHeader(statusCode)

	_, err := c.resp.Write([]byte(msg))

	return err
}
