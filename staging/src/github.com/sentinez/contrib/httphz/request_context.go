package httphz

import (
	"context"
	"io"
	"time"

	"github.com/a-h/templ"
	"github.com/cloudwego/hertz/pkg/app"
	corehttp "github.com/sentinez/core/http"
	httpconst "github.com/sentinez/core/http/const"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/dmz/edge/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/shared/sync"
)

var (
	_       corehttp.Context = (*Context)(nil)
	ctxPool                  = sync.NewPoolCtr(func() *Context {
		return &Context{
			x:       &edgepb.Context{},
			request: &typepb.Request{},
		}
	})
)

func NewContext(ctx context.Context, c *app.RequestContext) *Context {
	httpCtx := ctxPool.Get()

	httpCtx.req = c
	httpCtx.ctx = ctx

	return httpCtx
}

type Context struct {
	req *app.RequestContext
	ctx context.Context

	request *typepb.Request
	x       *edgepb.Context
}

// SetRequestId implements corehttp.Context.
func (c *Context) SetRequestId(id string) {
	if id == "" {
		return
	}

	c.request.Id = id
	c.req.Request.SetHeader(httpconst.HeaderXRequestId, id)
}

// Extra implements corehttp.Context.
func (c *Context) Extra() *edgepb.Context {
	return c.x
}

// SetExtra implements corehttp.Context.
func (c *Context) SetExtra(x *edgepb.Context) {
	c.x = x
}

func (c *Context) AddResponseHeader(key, value []byte) {
	c.req.Response.Header.Add(bytesToString(key), bytesToString(value))
}

func (c *Context) Flush() error {
	return c.req.Flush()
}

func (c *Context) Header(k []byte) []byte {
	return c.req.GetHeader(bytesToString(k))
}

func (c *Context) Query(k []byte) []byte {
	return []byte(c.req.Query(bytesToString(k)))
}

func (c *Context) RemoteAddr() []byte {
	return []byte(c.req.RemoteAddr().String())
}

func (c *Context) RequestId() string {
	return c.request.GetId()
}

func (c *Context) ResponseBody() []byte {
	return c.req.Response.Body()
}

func (c *Context) ResponseHeader() map[string][][]byte {
	headers := make(map[string][][]byte)
	c.req.Response.Header.VisitAll(func(key, value []byte) {
		headers[bytesToString(key)] = append(headers[bytesToString(key)], value)
	})

	return headers
}

func (c *Context) ResponseStatus() int {
	return c.req.Response.StatusCode()
}

func (c *Context) Scheme() string {
	return bytesToString(c.req.Request.Scheme())
}

func (c *Context) SetRequestIP(ip []byte) {
	c.request.ClientIp = bytesToString(ip)
}

func (c *Context) SetHeader(key, value []byte) {
	c.req.Request.Header.Set(bytesToString(key), bytesToString(value))
}

func (c *Context) SetHost(h []byte) {
	c.req.Request.SetHost(bytesToString(h))
}

func (c *Context) SetJA4(fingerprint string) {
	c.request.Fingerprint = fingerprint
}

func (c *Context) SetMethod(method []byte) {
	c.req.Request.SetMethod(bytesToString(method))
}

func (c *Context) SetProtocol(proto string) {
	c.req.Request.Header.SetProtocol(proto)
}

func (c *Context) SetQuery(key []byte, values ...[]byte) {
	for _, v := range values {
		c.req.Request.URI().QueryArgs().
			Add(bytesToString(key), bytesToString(v))
	}
}

func (c *Context) SetRemoteAddr(_ []byte) {
	c.req.RemoteAddr()
}

func (c *Context) SetResponseHeader(key, value []byte) {
	c.req.Response.Header.Set(bytesToString(key), bytesToString(value))
}

func (c *Context) SetResponseStatus(code int) {
	c.req.SetStatusCode(code)
}

func (c *Context) SetURI(u []byte) {
	c.req.Request.SetRequestURI(bytesToString(u))
}

func (c *Context) QueryStr() []byte {
	return c.req.URI().QueryString()
}

func (c *Context) SetPath(path []byte) {
	c.req.URI().SetPath(bytesToString(path))
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

func (c *Context) URI() []byte {
	return []byte(c.req.URI().String())
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

func (c *Context) Headers() map[string][][]byte {
	headers := make(map[string][][]byte)
	c.req.VisitAllHeaders(func(key, value []byte) {
		headers[bytesToString(key)] = append(headers[bytesToString(key)], value)
	})

	return headers
}

func (c *Context) Host() []byte {
	return c.req.Request.Host()
}

func (c *Context) RequestIP() []byte {
	return []byte(c.req.ClientIP())
}

func (c *Context) JA4() string {
	return c.request.GetFingerprint()
}

func (c *Context) Method() []byte {
	return c.req.Request.Method()
}

func (c *Context) Queries() map[string][][]byte {
	params := make(map[string][][]byte)

	c.req.VisitAllQueryArgs(func(key, value []byte) {
		k := bytesToString(key)
		params[k] = append(params[k], value)
	})

	return params
}

func (c *Context) TLS() bool {
	return c.Scheme() == httpconst.SchemeSecure
}

func (c *Context) RequestTime() time.Time {
	return corehttp.GetRequestTime(c.ctx)
}

func (c *Context) Context() context.Context {
	c.req.GetConn()
	return c.ctx
}

func (c *Context) JSON(statusCode int, body []byte) error {
	c.req.SetContentType(httpconst.ValueAppJSON)
	c.req.SetStatusCode(statusCode)

	_, err := c.req.Write(body)
	return err
}

func (c *Context) File(path string) error {
	c.req.File(path)
	return nil
}

func (c *Context) Path() []byte {
	return c.req.Request.URI().PathOriginal()
}

func (c *Context) String(statusCode int, body []byte) error {
	c.req.SetContentType(httpconst.ValueTextPlain)
	c.req.SetStatusCode(statusCode)

	_, err := c.req.Write(body)
	return err
}

func (c *Context) Render(statusCode int, component templ.Component) error {
	c.req.SetStatusCode(statusCode)
	c.req.SetContentType(httpconst.ValueTextHTML)

	return component.Render(c.Context(), c.req.Response.BodyWriter())
}

func (c *Context) Unwrap() any {
	return c.req
}

func (c *Context) Release() {
	c.req = nil
	c.ctx = nil

	c.x.Reset()
	c.request.Reset()

	ctxPool.Put(c)
}
