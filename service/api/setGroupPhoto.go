package api

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
	userId, err := strconv.Atoi(params.ByName("userId"))
	if err != nil {
		context.Logger.WithError(err).Error("Error in getGroup")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	groupId, err := strconv.Atoi(params.ByName("groupId"))
	if err != nil {
		context.Logger.WithError(err).Error("Error in getGroup")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")

	isThere, err := rt.db.UserGroup(userId, groupId, timestamp)
	if err != nil {
		context.Logger.WithError(err).Error("Error in getGroup")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !isThere {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	const maxSize = 10 << 20 // 10MB
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := r.ParseMultipartForm(maxSize); err != nil {
		context.Logger.WithError(err).Error("Error parsing multipart form")
		http.Error(w, "Invalid multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		context.Logger.WithError(err).Error("Error getting file from form")
		http.Error(w, "Photo not privided", http.StatusBadRequest)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			context.Logger.WithError(err).Error("Error closing file")
			http.Error(w, "File close error", http.StatusInternalServerError)
		}
	}(file)

	mime := header.Header.Get("Content-Type")
	if mime != "image/jpeg" && mime != "image/png" {
		context.Logger.WithError(err).Errorf("Invalid image type: %s", mime)
		http.Error(w, "Invalid image type", http.StatusBadRequest)
		return
	}

	filename := fmt.Sprintf("group_%d_%s.png", groupId, context.ReqUUID.String())
	path := filepath.Join("uploads", "groups", strconv.Itoa(groupId), filename)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		context.Logger.WithError(err).Error("Error creating directory")
		http.Error(w, "Cannot create directory", http.StatusInternalServerError)
		return
	}

	out, err := os.Create(path)
	if err != nil {
		context.Logger.WithError(err).Error("Error creating file")
		http.Error(w, "Cannot save image", http.StatusInternalServerError)
		return
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			context.Logger.WithError(err).Error("Error closing file")
			http.Error(w, "Cannot save image", http.StatusInternalServerError)
		}
	}(out)

	if _, err := io.Copy(out, file); err != nil {
		context.Logger.WithError(err).Error("Error copying file")
		http.Error(w, "Write failed", http.StatusInternalServerError)
		return
	}

	imgFile, err := os.Open(path)
	if err != nil {
		context.Logger.WithError(err).Error("Error opening file")
		http.Error(w, "Cannot open image", http.StatusInternalServerError)
		return
	}
	defer func(imgFile *os.File) {
		err := imgFile.Close()
		if err != nil {
			context.Logger.WithError(err).Error("Error closing file")
			http.Error(w, "Write failed", http.StatusInternalServerError)
		}
	}(imgFile)

	img, _, err := image.Decode(imgFile)
	if err != nil {
		context.Logger.WithError(err).Error("Error decoding jpeg")
		http.Error(w, fmt.Sprintf("Cannot decode JPEG: %v", err), http.StatusBadRequest)
		return
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	err = rt.db.SetGroupPhoto(path, width, height, mime, groupId)
	if err != nil {
		context.Logger.WithError(err).Error("Error setting photo")
		http.Error(w, "error while saving photo", http.StatusInternalServerError)
		return
	}

	resp := Upload{
		Url:    "/" + path,
		Mime:   mime,
		Width:  width,
		Height: height,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		context.Logger.WithError(err).Error("Error in getGroupPhoto")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
