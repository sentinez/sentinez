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
	"testing"
	"testing/fstest"

	"github.com/sentinez/sentinez"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestImages(t *testing.T) map[string]*image {
	t.Helper()

	images, err := loadImages(fstest.MapFS{
		"public/sntz.png":      {Data: []byte("png")},
		"public/logo/dark.SVG": {Data: []byte("<svg/>")},
		"public/photo.jpeg":    {Data: []byte("jpeg")},
		"public/notes.txt":     {Data: []byte("text")},
		"public/.gitkeep":      {},
	}, "public")
	require.NoError(t, err)
	require.Len(t, images, 3)

	return images
}

func TestLoadImages(t *testing.T) {
	images := newTestImages(t)

	tests := []struct {
		name            string
		giveKey         string
		wantBody        string
		wantContentType string
		wantLength      string
	}{
		{"root image", "sntz.png", "png", "image/png", "3"},
		{"nested upper-case ext", "logo/dark.SVG", "<svg/>",
			"image/svg+xml", "6"},
		{"jpeg", "photo.jpeg", "jpeg", "image/jpeg", "4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, ok := images[tt.giveKey]
			require.True(t, ok)

			assert.Equal(t, tt.wantBody, string(img.body))
			assert.Equal(t, tt.wantContentType, string(img.contentType))
			assert.Equal(t, tt.wantLength, string(img.length))
			assert.Regexp(t, `^"[0-9a-f]{16}"$`, string(img.etag))
		})
	}
}

func TestLoadImagesMissingRoot(t *testing.T) {
	_, err := loadImages(fstest.MapFS{}, "public")
	require.Error(t, err)
}

func TestLoadImagesEmbedded(t *testing.T) {
	images, err := loadImages(sentinez.Logo, _publicRoot)
	require.NoError(t, err)
	assert.Contains(t, images, "sntz.png")
}
