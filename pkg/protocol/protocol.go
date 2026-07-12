package protocol

import (
	"fmt"

	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/dmz/edge/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
)

func Upstream2Target(upstream *edgepb.Upstream) (string, error) {
	var target string
	switch upstream.GetProtocol() {
	case edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTP:
		target = "http://" + upstream.GetServer()
	case edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTPS:
		target = "https://" + upstream.GetServer()
	default:
		return "",
			fmt.Errorf("unsupported protocol: %v", upstream.GetProtocol())
	}

	return target, nil
}

func ParseQuery[S []string | [][]byte](
	src map[string]S, dst map[string]*typepb.QueryValue) {

	if dst == nil {
		return
	}

	for k, v := range src {
		q, ok := dst[k]
		if !ok {
			q = &typepb.QueryValue{
				Values: make([]string, 0, len(v)),
			}
		}

		switch values := any(v).(type) {
		case []string:
			q.Values = append(q.Values, values...)
			dst[k] = q
		case [][]byte:

			for _, b := range values {
				q.Values = append(q.Values, string(append([]byte(nil), b...)))
			}

			dst[k] = q
		}
	}
}

func ParseHeader[S []string | [][]byte](
	src map[string]S, dst map[string]*typepb.HeaderValue) {

	if dst == nil {
		return
	}

	for k, v := range src {
		q, ok := dst[k]
		if !ok {
			q = &typepb.HeaderValue{
				Values: make([]string, 0, len(v)),
			}
		}

		switch values := any(v).(type) {
		case []string:
			q.Values = append(q.Values, values...)
			dst[k] = q
		case [][]byte:

			for _, b := range values {
				q.Values = append(q.Values, string(append([]byte(nil), b...)))
			}

			dst[k] = q
		}
	}
}
