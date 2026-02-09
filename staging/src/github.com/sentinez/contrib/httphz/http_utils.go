package httphz

import (
	"context"
	"unsafe"

	"github.com/cloudwego/hertz/pkg/app"
	corehttp "github.com/sentinez/core/http"
)

func WrapHandler(next app.HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {

		ctx = corehttp.SetRequestTime(ctx)

		next(ctx, c)
	}
}

func bytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}
