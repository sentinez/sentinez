package proxyhz

import (
	"bytes"
	"fmt"
	"strings"
	"unsafe"

	"github.com/cloudwego/hertz/pkg/protocol"
)

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// nolint:funlen
func JoinURLPath(req *protocol.Request, target string) (path []byte) {
	aslash := req.URI().Path()[0] == '/'
	var bslash bool
	if strings.HasPrefix(target, "http") {
		// absolute path
		bslash = strings.HasSuffix(target, "/")
	} else {
		// default redirect to local
		bslash = strings.HasPrefix(target, "/")
		if bslash {
			target = fmt.Sprintf("%s%s", req.Host(), target)
		} else {
			target = fmt.Sprintf("%s/%s", req.Host(), target)
		}
		bslash = strings.HasSuffix(target, "/")
	}

	targetQuery := strings.Split(target, "?")
	var buffer bytes.Buffer
	buffer.WriteString(targetQuery[0])
	switch {
	case aslash && bslash:
		buffer.Write(req.URI().Path()[1:])
	case !aslash && !bslash:
		buffer.Write([]byte{'/'})
		buffer.Write(req.URI().Path())
	default:
		buffer.Write(req.URI().Path())
	}
	if len(targetQuery) > 1 {
		buffer.Write([]byte{'?'})
		buffer.WriteString(targetQuery[1])
	}
	if len(req.QueryString()) > 0 {
		if len(targetQuery) == 1 {
			buffer.Write([]byte{'?'})
		} else {
			buffer.Write([]byte{'&'})
		}
		buffer.Write(req.QueryString())
	}
	return buffer.Bytes()
}
