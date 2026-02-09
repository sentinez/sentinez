package proxyhz

import (
	"crypto/tls"
	"time"
)

// Option to define all options to reverse http proxy.
type Option interface {
	apply(o *buildOption)
}

// buildOption contains all fields those are used in ReverseProxy.
type buildOption struct {

	// debug to open debug mode to log more info to logger
	debug bool

	// openBalance denote whether the balancer is configured or not.
	openBalance bool

	// tlsConfig is pointer to tls.Config, will be used if the upstream.
	// need TLS handshake
	tlsConfig *tls.Config

	// timeout specify the timeout context with each request.
	timeout time.Duration

	// disablePathNormalizing disable path normalizing.
	disablePathNormalizing bool

	// disableVirtualHost disable virtual host.
	disableVirtualHost bool

	// maxResponseBodySize is the maximum response body size in bytes.
	maxResponseBodySize int

	// streamResponseBody denotes whether stream response body or not.
	streamResponseBody bool

	// maxConnDuration of hostClient
	maxConnDuration time.Duration
}

func defaultBuildOption() *buildOption {
	return &buildOption{
		debug:                  false,
		openBalance:            false,
		tlsConfig:              nil,
		timeout:                0,
		disablePathNormalizing: false,
		disableVirtualHost:     false,
		maxConnDuration:        0,
	}
}

type funcBuildOption struct {
	f func(o *buildOption)
}

func newFuncBuildOption(f func(o *buildOption)) funcBuildOption {
	return funcBuildOption{f: f}
}

func (fb funcBuildOption) apply(o *buildOption) { fb.f(o) }

func WithTLSConfig(config *tls.Config) Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.tlsConfig = config
	})
}

// WithTLS build tls.Config with server certFile and keyFile.
// tlsConfig is nil as default
func WithTLS(certFile, keyFile string) Option {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		panic("" + err.Error())
	}

	return WithTLSConfig(&tls.Config{
		Certificates: []tls.Certificate{cert},
	})
}

func WithDebug() Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.debug = true
	})
}

// WithTimeout specify the timeout of each request
func WithTimeout(d time.Duration) Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.timeout = d
	})
}

// WithDisablePathNormalizing sets whether disable path normalizing.
func WithDisablePathNormalizing(isDisablePathNormalizing bool) Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.disablePathNormalizing = isDisablePathNormalizing
	})
}

// WithDisableVirtualHost sets whether disable virtual host.
func WithDisableVirtualHost(isDisableVirtualHost bool) Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.disableVirtualHost = isDisableVirtualHost
	})
}

// WithStreamResponseBody sets whether stream response body or not.
func WithStreamResponseBody(size int) Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.streamResponseBody = true
		o.maxResponseBodySize = size
	})
}

// WithMaxConnDuration sets maxConnDuration of hostClient, which
// means keep-alive connections are closed after this duration.
func WithMaxConnDuration(d time.Duration) Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.maxConnDuration = d
	})
}
