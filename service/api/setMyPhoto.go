package api

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type Upload struct {
	Url    string `json:"url"`
	Mime   string `json:"mime"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {

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

	userId := params.ByName("userId")
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		context.Logger.WithError(err).Errorf("Invalid user id: %s", userId)
		http.Error(w, "Invalid user id", http.StatusBadRequest)
	}
	filename := fmt.Sprintf("avatar_%s_%s.png", userId, context.ReqUUID.String())
	path := filepath.Join("uploads", "users", userId, filename)

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

	var img image.Image

	switch mime {
	case "image/jpeg":
		img, err = jpeg.Decode(imgFile)
		if err != nil {
			context.Logger.WithError(err).Error("Error decoding jpeg")
			http.Error(w, fmt.Sprintf("Cannot decode JPEG: %v", err), http.StatusBadRequest)
			return
		}
	case "image/png":
		img, err = png.Decode(imgFile)
		if err != nil {
			context.Logger.WithError(err).Error("Error decoding png")
			http.Error(w, fmt.Sprintf("Cannot decode PNG: %v", err), http.StatusBadRequest)
			return
		}
	default:
		context.Logger.WithError(err).Errorf("Invalid image type: %s", mime)
		http.Error(w, "Invalid image type", http.StatusBadRequest)
		return
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	err = rt.db.SetMyPhoto(path, width, height, mime, userIdInt)
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
		context.Logger.WithError(err).Error("Error encoding response")
		http.Error(w, "error while saving response", http.StatusInternalServerError)
		return
	}

}
