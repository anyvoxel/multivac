// Package webui serves the embedded React web UI.
package webui

import (
	"embed"
	"errors"
	"io/fs"
	"mime"
	"path"
	"path/filepath"
	"strings"
)

// distFS contains the built web assets copied from `web/dist`.
//
// NOTE: build-time flow:
//  1. `cd web && npm run build` generates `web/dist`
//  2. `make build` copies it into `internal/webui/dist`
//  3. this embed packs files into the binary
//
//go:embed dist/* dist/assets/*
var distFS embed.FS

// Asset is a static file payload read from embedded dist.
type Asset struct {
	Path        string
	ContentType string
	Body        []byte
	Immutable   bool
}

// ReadAsset returns embedded SPA assets for the given request path.
func ReadAsset(reqPath string) (Asset, error) {
	p := normalizePath(reqPath)
	if p == "/" || p == "" {
		return readFile("dist/index.html", "text/html; charset=utf-8", false)
	}

	// Serve static assets.
	if strings.HasPrefix(p, "/assets/") {
		return readByPath(p[1:], true)
	}
	if p == "/favicon.svg" {
		return readByPath(p[1:], true)
	}
	if p == "/vite.svg" {
		return readByPath(p[1:], true)
	}

	// SPA fallback.
	return readFile("dist/index.html", "text/html; charset=utf-8", false)
}

func readByPath(rel string, immutable bool) (Asset, error) {
	rel = path.Clean("/" + rel)
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" || strings.Contains(rel, "..") {
		return Asset{}, fs.ErrNotExist
	}

	contentType := mime.TypeByExtension(filepath.Ext(rel))
	if contentType == "" {
		switch filepath.Ext(rel) {
		case ".js":
			contentType = "application/javascript; charset=utf-8"
		case ".css":
			contentType = "text/css; charset=utf-8"
		case ".svg":
			contentType = "image/svg+xml"
		}
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return readFile("dist/"+rel, contentType, immutable)
}

func readFile(embedPath, contentType string, immutable bool) (Asset, error) {
	b, err := distFS.ReadFile(embedPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Asset{}, fs.ErrNotExist
		}
		return Asset{}, err
	}
	return Asset{Path: embedPath, ContentType: contentType, Body: b, Immutable: immutable}, nil
}

func normalizePath(p string) string {
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return path.Clean(p)
}
