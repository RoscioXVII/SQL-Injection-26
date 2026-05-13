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

	http.ServeFile(w, r, filePath)
}
