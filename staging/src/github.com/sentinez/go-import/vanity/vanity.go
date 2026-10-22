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

// Package vanity serves go-import/go-source meta tags so that vanity
// import paths (go.sentinez.com/<repo>/...) resolve to GitHub repos.
// Requests without ?go-get=1 get a human-readable project page instead.
package vanity

import (
	"embed"
	"net/http"
	"slices"
	"strings"

	"github.com/a-h/templ"
)

//go:generate go run github.com/a-h/templ/cmd/templ@v0.3.1020 generate

// Config describes the vanity host and the repos it serves.
type Config struct {
	// Host is the vanity domain, e.g. "go.sentinez.com".
	Host string
	// Org is the base URL of the repos, e.g. "https://github.com/sentinez".
	Org string
	// Branch is the default branch used for go-source links.
	Branch string
	// Project is the project name shown on info pages.
	Project string
	// Description is the project summary shown on the index page.
	Description string
	// Repos is the allowlist of repos served under Host.
	Repos []Repo
}

// Repo is a repo served under the vanity host.
type Repo struct {
	// Name is both the first import path segment and the repo name.
	Name string
	// Description is a one-line summary shown on info pages.
	Description string
}

// Handler answers `go get` meta requests for allowlisted repos and
// serves project info pages to browsers.
type Handler struct {
	conf   Config
	static http.Handler
}

type metaPage struct {
	Root   string
	Repo   string
	Branch string
}

type infoPage struct {
	Project     string
	Description string
	// Path is the requested import path; empty on the index page.
	Path  string
	Repos []repoView
}

type repoView struct {
	Name        string
	Description string
	Import      string
	Repo        string
	Docs        string
}

var _ http.Handler = (*Handler)(nil)

const _staticPrefix = "/_static/"

// _staticFS holds the logo images served under _staticPrefix.
//
//go:embed static/*.png
var _staticFS embed.FS

// DefaultConfig returns the config for go.sentinez.com.
func DefaultConfig() Config {
	return Config{
		Host:    "go.sentinez.com",
		Org:     "https://github.com/sentinez",
		Branch:  "main",
		Project: "Sentinéz",
		Description: "Go security platform: WAF / edge proxy with " +
			"OWASP CRS rule-based filtering, rate limiting " +
			"and a web console.",
		Repos: _defaultRepos(),
	}
}

// New returns a Handler for conf.
func New(conf Config) *Handler {
	conf.Repos = slices.Clone(conf.Repos)
	// "/_static/logo.png" -> "static/logo.png" in _staticFS.
	static := http.StripPrefix("/_", http.FileServerFS(_staticFS))
	return &Handler{conf: conf, static: static}
}

// ServeHTTP writes the go-import meta page for ?go-get=1 requests and
// an info page otherwise. The repo is the first path segment, e.g.
// /core/szhttp/chains -> github.com/sentinez/core.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if strings.HasPrefix(r.URL.Path, _staticPrefix) {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		h.static.ServeHTTP(w, r)
		return
	}

	path := strings.Trim(r.URL.Path, "/")
	if path == "" {
		render(w, r, infoView(h.index()))
		return
	}

	name, _, _ := strings.Cut(path, "/")
	idx := slices.IndexFunc(h.conf.Repos, func(repo Repo) bool {
		return repo.Name == name
	})
	if idx < 0 {
		http.NotFound(w, r)
		return
	}

	repo := h.conf.Repos[idx]
	if r.URL.Query().Get("go-get") == "1" {
		render(w, r, metaView(h.meta(repo)))
		return
	}
	render(w, r, infoView(h.info(repo, path)))
}

func render(w http.ResponseWriter, r *http.Request, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	if err := c.Render(r.Context(), w); err != nil {
		http.Error(w, "render page", http.StatusInternalServerError)
	}
}

// source is the go-source meta content: prefix, home, dir and file URLs.
func (p metaPage) source() string {
	tree := p.Repo + "/tree/" + p.Branch
	blob := p.Repo + "/blob/" + p.Branch
	return p.Root + " " + p.Repo + " " + tree + "{/dir} " +
		blob + "{/dir}/{file}#L{line}"
}

func (p infoPage) title() string {
	if p.Path != "" {
		return p.Path
	}
	return p.Project + " Go modules"
}

func (h *Handler) meta(repo Repo) metaPage {
	return metaPage{
		Root:   h.conf.Host + "/" + repo.Name,
		Repo:   h.conf.Org + "/" + repo.Name,
		Branch: h.conf.Branch,
	}
}

func (h *Handler) index() infoPage {
	views := make([]repoView, 0, len(h.conf.Repos))
	for _, repo := range h.conf.Repos {
		views = append(views, h.view(repo, repo.Name))
	}
	return infoPage{
		Project:     h.conf.Project,
		Description: h.conf.Description,
		Repos:       views,
	}
}

func (h *Handler) info(repo Repo, path string) infoPage {
	return infoPage{
		Project:     h.conf.Project,
		Description: h.conf.Description,
		Path:        h.conf.Host + "/" + path,
		Repos:       []repoView{h.view(repo, path)},
	}
}

func (h *Handler) view(repo Repo, path string) repoView {
	return repoView{
		Name:        repo.Name,
		Description: repo.Description,
		Import:      h.conf.Host + "/" + path,
		Repo:        h.conf.Org + "/" + repo.Name,
		Docs:        "https://pkg.go.dev/" + h.conf.Host + "/" + path,
	}
}

func _defaultRepos() []Repo {
	return []Repo{
		{"sentinez", "Main repo: services, edge proxy, API server."},
		{"core", "Runtime framework: runner, HTTP/gRPC, rate " +
			"limiter, ModSecurity (Coraza), rules, CRS."},
		{"shared", "Utility libraries: logging, config, errors, " +
			"event queue, store, cron."},
		{"controlplane", "Domain services: IAM, tenant, security, " +
			"analytic, central data."},
		{"contrib", "Contributed helpers (httphz)."},
		{"bpf", "eBPF programs (cilium/ebpf)."},
		{"tools", "Code generation and rule parser tooling."},
		{"mcp", "Model Context Protocol integration."},
		{"go-import", "Vanity import server for go.sentinez.com."},
	}
}
