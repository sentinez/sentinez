// Copyright 2026 Duc-Hung Ho.
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

package cdn

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/sentinez/core/szhttp/chains"

	"github.com/sentinez/core/szbytestr"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/httpconst"
	"github.com/sentinez/sentinez"
	"github.com/sentinez/sentinez/internal/envplugin"
)

// PublicPrefix is the path the embedded images are served under, e.g.
// _logo/sentinez/1/1.png is served at /cdn/sentinez/public/sentinez/1/1.png.
var PublicPrefix = httpconst.SystemPath("public/")

// _publicRoot is the directory sentinez.Logo embeds; paths inside the
// embed.FS keep it as their prefix.
const _publicRoot = "_logo"

var (
	_publicPrefix       = []byte(PublicPrefix)
	_publicCacheControl = []byte("public, max-age=86400")
)

// _imageTypes lists the only file types served; anything else in the
// embedded directory (LICENSE, README, ...) is ignored.
var _imageTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".avif": "image/avif",
	".svg":  "image/svg+xml",
	".ico":  "image/x-icon",
}

// PublicId is the chain stage name of Public.
const PublicId = "PUB"

var (
	_ chains.Plugin                      = (*Public)(nil)
	_ chains.Provisioner[*envplugin.Env] = (*Public)(nil)
)

// Public serves the images embedded in the binary under PublicPrefix.
// Requests outside the prefix are passed to the next plugin untouched.
type Public struct {
	// images is keyed by the path relative to PublicPrefix and is
	// read-only after provisioning.
	images map[string]*image
}

type image struct {
	body        []byte
	contentType []byte
	length      []byte
	etag        []byte
}

func (*Public) ID() string { return PublicId }

func (p *Public) Provision(*envplugin.Env) error {
	images, err := loadImages(sentinez.Logo, _publicRoot)
	if err != nil {
		return fmt.Errorf("load public images: %w", err)
	}

	p.images = images

	return nil
}

func (p *Public) Handle(ctx szhttp.Context, next chains.Handler) error {
	name, ok := bytes.CutPrefix(ctx.Path(), _publicPrefix)
	if !ok {
		return next.Handle(ctx)
	}

	method := string(ctx.Method())
	if method != http.MethodGet && method != http.MethodHead {
		ctx.SetResponseHeader(
			szbytestr.HeaderAllow, szbytestr.ValueAllowGetHead)
		return ctx.String(
			http.StatusMethodNotAllowed, szbytestr.MethodNotAllowed)
	}

	img, ok := p.images[string(name)]
	if !ok {
		return szhttp.NotFound(ctx)
	}

	return serveImage(ctx, img)
}

func serveImage(ctx szhttp.Context, img *image) error {
	ctx.SetResponseHeader(szbytestr.HeaderServer, szbytestr.DefaultServerName)
	ctx.SetResponseHeader(szbytestr.HeaderCacheControl, _publicCacheControl)
	ctx.SetResponseHeader(szbytestr.HeaderETag, img.etag)

	if bytes.Contains(ctx.Header(szbytestr.HeaderIfNoneMatch), img.etag) {
		ctx.SetStatusCode(http.StatusNotModified)
		return nil
	}

	ctx.SetResponseHeader(szbytestr.HeaderContentType, img.contentType)
	ctx.SetResponseHeader(szbytestr.HeaderContentLength, img.length)
	ctx.SetResponseHeader(
		szbytestr.HeaderXContentTypeOpts, szbytestr.ValueNoSniff)
	// Status must be written before the body, or it is silently dropped.
	ctx.SetStatusCode(http.StatusOK)
	ctx.SetBody(img.body)

	return nil
}

// loadImages reads every image under root into memory, keyed by its path
// relative to root.
func loadImages(fsys fs.FS, root string) (map[string]*image, error) {
	images := make(map[string]*image)

	err := fs.WalkDir(fsys, root,
		func(name string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}

			ct, ok := _imageTypes[strings.ToLower(path.Ext(name))]
			if !ok {
				return nil
			}

			body, err := fs.ReadFile(fsys, name)
			if err != nil {
				return fmt.Errorf("read %q: %w", name, err)
			}

			images[strings.TrimPrefix(name, root+"/")] = newImage(body, ct)

			return nil
		})
	if err != nil {
		return nil, err
	}

	return images, nil
}

func newImage(body []byte, contentType string) *image {
	sum := sha256.Sum256(body)

	return &image{
		body:        body,
		contentType: []byte(contentType),
		length:      []byte(strconv.Itoa(len(body))),
		etag:        []byte(`"` + hex.EncodeToString(sum[:8]) + `"`),
	}
}
