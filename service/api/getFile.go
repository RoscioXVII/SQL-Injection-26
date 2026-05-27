package api

import (
	"net/http"
	"path/filepath"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getFile(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
	fileName := r.URL.Query().Get("file")

	if fileName == "" {
		http.Error(w, "file parameter missing", http.StatusBadRequest)
		return
	}

	filePath := filepath.Clean(fileName)
	if !(strings.HasPrefix(filePath, "uploads") || strings.HasPrefix(filePath, "assets")) {
		http.Error(w, "invalid file path", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox") // !!
	w.Header().Set("X-Content-Type-Options", "nosniff")                                                 // !!
	http.ServeFile(w, r, filePath)
}
