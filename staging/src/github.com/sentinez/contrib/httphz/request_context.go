package httphz

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/a-h/templ"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gorilla/websocket"
	corehttp "github.com/sentinez/core/http"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/zlog"
)

var (
	_       corehttp.Context = (*Context)(nil)
	ctxPool                  = sync.NewPool[Context]()
)

func NewContext(ctx context.Context, c *app.RequestContext) *Context {

	httpCtx := ctxPool.Get()

	httpCtx.req = c
	httpCtx.ctx = ctx

	return httpCtx
}

type Context struct {
	id  string
	req *app.RequestContext
	ctx context.Context
	x   *edgepb.Context
}

// SetRequestId implements corehttp.Context.
func (c *Context) SetRequestId(id string) {
	c.id = id
}

// Extra implements corehttp.Context.
func (c *Context) Extra() *edgepb.Context {
	return c.x
}

// SetExtra implements corehttp.Context.
func (c *Context) SetExtra(x *edgepb.Context) {
	c.x = x
}

func (c *Context) AddResponseHeader(key string, value string) {
	c.req.Response.Header.Add(key, value)
}

func (c *Context) Flush() error {
	return c.req.Flush()
}

func (c *Context) Header(k string) string {
	return bytesToString(c.req.GetHeader(k))
}

func (c *Context) Query(k string) string {
	return c.req.Query(k)
}

func (c *Context) RemoteAddr() string {
	return c.req.RemoteAddr().String()
}

func (c *Context) RequestId() string {
	return c.Header(corehttp.HeaderXRequestId)
}

func (c *Context) ResponseBody() []byte {
	return c.req.Response.Body()
}

func (c *Context) ResponseHeader() map[string]string {
	headers := make(map[string]string)
	c.req.Response.Header.VisitAll(func(key, value []byte) {
		headers[(bytesToString(key))] = bytesToString(value)
	})

	return headers
}

func (c *Context) ResponseStatus() int {
	return c.req.Response.StatusCode()
}

func (c *Context) Scheme() string {
	return bytesToString(c.req.Request.Scheme())
}

func (c *Context) SetRequestIP(_ string) {
	zlog.Fatal("[httpxdmz] unimplemented")
}

func (c *Context) SetHeader(key string, value string) {
	c.req.Request.Header.Set(key, value)
}

func (c *Context) SetHost(h string) {
	c.req.Request.SetHost(h)
}

func (c *Context) SetJA4(_ string) {
	zlog.Fatal("[httpxdmz] unimplemented")
}

func (c *Context) SetMethod(method string) {
	c.req.Request.SetMethod(method)
}

func (c *Context) SetProtocol(proto string) {
	c.req.Request.Header.SetProtocol(proto)
}

func (c *Context) SetQuery(_ string, _ ...string) {
	zlog.Fatal("[httpxdmz] unimplemented")
}

func (c *Context) SetRemoteAddr(_ string) {
	zlog.Fatal("[httpxdmz] unimplemented")
}

func (c *Context) SetResponseHeader(key string, value string) {
	c.req.Response.Header.Set(key, value)
}

func (c *Context) SetResponseStatus(code int) {
	c.req.SetStatusCode(code)
}

func (c *Context) SetURI(u string) {
	c.req.Request.SetRequestURI(u)
}

func (c *Context) Upgrade() (*websocket.Conn, error) {
	zlog.Fatal("[httpxdmz] unimplemented")
	return nil, errors.New("unimplemented")
}

func (c *Context) QueryStr() string {
	return c.req.QueryArgs().String()
}

func (c *Context) SetPath(path string) {
	c.req.URI().SetPath(path)
}

func (c *Context) Copy(src io.Reader) error {
	_, err := io.Copy(c.req, src)
	return err
}

func (c *Context) RequestBodyStream() io.Reader {
	return c.req.RequestBodyStream()
}

func (c *Context) VisitResponseHeaders(visitor func(k []byte, v []byte)) {
	c.req.Response.Header.VisitAll(func(key, value []byte) {
		visitor(key, value)
	})
}

func (c *Context) Protocol() string {
	return c.req.Request.Header.GetProtocol()
}

func (c *Context) ResetResponse() {
	c.req.Response.Reset()
}

func (c *Context) SetBody(body []byte) {
	c.req.Response.SetBody(body)
}

func (c *Context) SetStatusCode(code int) {
	c.req.SetStatusCode(code)
}

func (c *Context) StatusCode() int {
	return c.req.Response.StatusCode()
}

func (c *Context) URI() string {
	return c.req.URI().String()
}

func (c *Context) VisitRequestHeaders(visitor func(k []byte, v []byte)) {
	c.req.VisitAllHeaders(func(key, value []byte) {
		visitor(key, value)
	})
}

func (c *Context) Body() []byte {
	return c.req.Request.Body()
}

func (c *Context) GetContext() context.Context {
	return c.Context()
}

func (c *Context) Headers() map[string]string {
	headers := make(map[string]string)
	c.req.VisitAllHeaders(func(key, value []byte) {
		headers[(bytesToString(key))] = bytesToString(value)
	})

	return headers
}

func (c *Context) Host() string {
	return bytesToString(c.req.Request.Host())
}

func (c *Context) RequestIP() string {
	return c.req.ClientIP()
}

func (c *Context) JA4() string {
	return ""
}

func (c *Context) Method() string {
	return bytesToString(c.req.Request.Method())
}

func (c *Context) Queries() map[string][]string {
	params := make(map[string][]string)

	c.req.VisitAllQueryArgs(func(key, value []byte) {
		k := bytesToString(key)
		v := bytesToString(value)
		params[k] = append(params[k], v)
	})

	return params
}

func (c *Context) TLS() bool {
	return c.Scheme() == corehttp.SchemeSecure
}

func (c *Context) RequestTime() time.Time {
	return corehttp.GetRequestTime(c.ctx)
}

func (c *Context) Context() context.Context {
	c.req.GetConn()
	return c.ctx
}

func (c *Context) JSON(statusCode int, body []byte) error {
	c.req.SetContentType("application/json")
	c.req.SetStatusCode(statusCode)
	// c.SetServer(sentinez.Name)

	_, err := c.req.Write(body)
	return err
}

func (c *Context) File(path string) error {
	c.req.File(path)
	return nil
}

func (c *Context) Path() string {
	return bytesToString(c.req.Request.URI().PathOriginal())
}

func (c *Context) String(statusCode int, body string) error {
	c.req.SetContentType(corehttp.ValueTextPlain)
	c.req.SetStatusCode(statusCode)
	// c.SetServer(sentinez.Name)

	_, err := c.req.WriteString(body)
	return err
}

func (c *Context) Render(statusCode int, component templ.Component) error {
	c.req.SetStatusCode(statusCode)
	c.req.SetContentType(corehttp.ValueTextHTML)
	// c.SetServer(sentinez.Name)

	return component.Render(c.Context(), c.req.Response.BodyWriter())
}

func (c *Context) Unwrap() any {
	return c.req
}

func (c *Context) SetServer(name string) {
	c.SetResponseHeader(corehttp.HeaderServer, name)
}

func (c *Context) Release() {
	c.req = nil
	c.ctx = nil
	c.x = nil
	ctxPool.Put(c)
}
