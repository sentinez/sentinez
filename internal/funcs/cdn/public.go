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

	"github.com/sentinez/core/common/bytestr"
	corehttp "github.com/sentinez/core/http"
	corechains "github.com/sentinez/core/http/chains"
	"github.com/sentinez/sentinez"
	"github.com/sentinez/sentinez/internal/memory"
	"github.com/sentinez/shared/zlog"
)

// PublicPrefix is the path the embedded images are served under, e.g.
// _logo/sentinez/1/1.png is served at /cdn/sentinez/public/sentinez/1/1.png.
const PublicPrefix = "/cdn/sentinez/public/"

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

var _ corechains.ChainNode = (*Public)(nil)

// NewPublic returns a node serving the images embedded in the binary.
func NewPublic(_ zlog.Level, _ *memory.MemStore) corechains.ChainNode {
	images, err := loadImages(sentinez.Logo, _publicRoot)
	if err != nil {
		zlog.Error("[edge] load public images: ", err)
	}

	return &Public{
		Node:   corechains.NewNode(),
		images: images,
	}
}

// Public serves embedded images under PublicPrefix. Requests outside the
// prefix are passed to the next node untouched.
type Public struct {
	*corechains.Node

	// images is keyed by the path relative to PublicPrefix and is
	// read-only after construction.
	images map[string]*image
}

type image struct {
	body        []byte
	contentType []byte
	length      []byte
	etag        []byte
}

func (p *Public) Handle(ctx corehttp.Context) error {
	name, ok := bytes.CutPrefix(ctx.Path(), _publicPrefix)
	if !ok {
		return p.HandleNext(ctx)
	}

	method := string(ctx.Method())
	if method != http.MethodGet && method != http.MethodHead {
		ctx.SetResponseHeader(
			bytestr.HeaderAllow, bytestr.ValueAllowGetHead)
		return ctx.String(
			http.StatusMethodNotAllowed, bytestr.MethodNotAllowed)
	}

	img, ok := p.images[string(name)]
	if !ok {
		return corehttp.NotFound(ctx)
	}

	return serveImage(ctx, img)
}

func serveImage(ctx corehttp.Context, img *image) error {
	ctx.SetResponseHeader(bytestr.HeaderServer, bytestr.DefaultServerName)
	ctx.SetResponseHeader(bytestr.HeaderCacheControl, _publicCacheControl)
	ctx.SetResponseHeader(bytestr.HeaderETag, img.etag)

	if bytes.Contains(ctx.Header(bytestr.HeaderIfNoneMatch), img.etag) {
		ctx.SetStatusCode(http.StatusNotModified)
		return nil
	}

	ctx.SetResponseHeader(bytestr.HeaderContentType, img.contentType)
	ctx.SetResponseHeader(bytestr.HeaderContentLength, img.length)
	ctx.SetResponseHeader(
		bytestr.HeaderXContentTypeOpts, bytestr.ValueNoSniff)
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
