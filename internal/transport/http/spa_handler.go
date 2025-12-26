package http

import (
	"errors"
	"io/fs"
	"net/http"
	"strings"
)

// SPAHandler serves a Single Page Application from a static filesystem.
// It serves static files if they exist, otherwise it falls back to index.html.
type SPAHandler struct {
	StaticFS fs.FS
}

func (h SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// r.URL.Path here is already stripped of the prefix if using http.StripPrefix
	path := strings.TrimPrefix(r.URL.Path, "/")

	// If path is empty, it means we are at root (e.g. /admin/), serve index.html
	if path == "" {
		h.serveIndex(w)
		return
	}

	// Check if the file exists in the static directory
	f, err := h.StaticFS.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// File does not exist, serve index.html for client-side routing
			h.serveIndex(w)
			return
		}
		// Some other error
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	// If it's a directory, we might want to serve index.html or let FileServer handle it (usually 404 or directory listing)
	// But in SPA, we usually want index.html for directory paths too if they are routes.
	// For now, let's treat directories as not found -> index.html?
	// fs.Open generally opens the directory.
	stat, err := f.Stat()
	if err == nil && stat.IsDir() {
		h.serveIndex(w)
		return
	}

	// File exists and is a file, serve it
	http.FileServer(http.FS(h.StaticFS)).ServeHTTP(w, r)
}

func (h SPAHandler) serveIndex(w http.ResponseWriter) {
	content, err := fs.ReadFile(h.StaticFS, "index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}
