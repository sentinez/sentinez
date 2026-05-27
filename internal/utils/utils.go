package utils

import (
	"fmt"

	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
)

func Upstream2Target(upstream *edgepb.Upstream) (string, error) {
	var target string
	switch upstream.GetProtocol() {
	case edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTP:
		target = "http://" + upstream.GetServer()
	case edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTPS:
		target = "https://" + upstream.GetServer()
	default:
		return "", fmt.Errorf(
			"edge: unsupported protocol: %v", upstream.GetProtocol())
	}

	return target, nil
}
