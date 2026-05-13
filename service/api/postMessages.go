package api

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type Send struct {
	Text  string `json:"text"`
	Photo Upload `json:"photo"`
}

const JPEG = "image/jpeg"
const PNG = "image/png"

func (rt *_router) postMessage(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
	userId, err := strconv.Atoi(params.ByName("userId"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if !rt.db.IDExists(userId) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	replyToStr := r.URL.Query().Get("replyTo")

	var replyTo *int

	if replyToStr != "" {
		v, err := strconv.Atoi(replyToStr)
		if err != nil {
			http.Error(w, "Invalid replyTo", http.StatusBadRequest)
			return
		}
		replyTo = &v
	}
	conversationId, err := strconv.Atoi(params.ByName("conversationId"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	isThere, err := rt.db.UserConversation(userId, conversationId)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !isThere {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	contentType := r.Header.Get("Content-Type")
	isMultipart := strings.HasPrefix(contentType, "multipart/")

	const maxSize = 10 << 20 // 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	var text string
	var photoId *int

	// -------- CASE 1: JSON SENZA FOTO --------
	if !isMultipart {
		var payload struct {
			Body string `json:"bodyText"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		text = payload.Body
	} else {
		// -------- CASE 2: MULTIPART (TESTO + FACOLTATIVA FOTO) --------
		if err := r.ParseMultipartForm(maxSize); err != nil {
			http.Error(w, "Invalid multipart form", http.StatusBadRequest)
			return
		}

		text = r.FormValue("bodyText")

		file, header, err := r.FormFile("photo")

		if err == nil {

			mime := header.Header.Get("Content-Type")

			if mime != JPEG && mime != PNG {
				http.Error(w, "Invalid image type", http.StatusBadRequest)
				return
			}

			ext := ".png"
			if mime == JPEG {
				ext = ".jpg"
			}

			defer func(file multipart.File) {
				err := file.Close()
				if err != nil {
					context.Logger.WithError(err).Error("closing file")
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
			}(file)

			path := fmt.Sprintf("uploads/conversations/%d_%d_%s%s",
				conversationId, userId, context.ReqUUID, ext,
			)

			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				http.Error(w, "Cannot create directory", http.StatusInternalServerError)
				return
			}

			out, err := os.Create(path)
			if err != nil {
				http.Error(w, "Cannot save image", http.StatusInternalServerError)
				context.Logger.WithError(err).Error("Cannot save image")
				return
			}
			defer func(out *os.File) {
				err := out.Close()
				if err != nil {
					http.Error(w, "Cannot save image", http.StatusInternalServerError)
					context.Logger.Errorf("Cannot save image: %s", err)
				}
			}(out)

			if _, err := io.Copy(out, file); err != nil {
				http.Error(w, "Write failed", http.StatusInternalServerError)
				return
			}
			imgFile, err := os.Open(path)
			if err != nil {
				context.Logger.WithError(err).WithField("path", path).Error("Cannot open file")
				http.Error(w, "Cannot save image", http.StatusInternalServerError)
				return
			}

			img, _, err := image.Decode(imgFile)

			if err != nil {
				context.Logger.WithError(err).Error("Error decoding photo file")
				http.Error(w, "Cannot decode image", http.StatusInternalServerError)
				return
			}
			width := img.Bounds().Dx()
			height := img.Bounds().Dy()
			id, err := rt.db.InsertPhoto(path, width, height, mime)
			if err != nil {
				http.Error(w, "Cannot save photo to DB", http.StatusInternalServerError)
				return
			}

			photoId = &id
		}
	}

	// ALMENO TESTO O FOTO
	if text == "" && photoId == nil {
		http.Error(w, "Message must have text or photo", http.StatusBadRequest)
		return
	}

	// SALVA MESSAGGIO
	mess, err := rt.db.InsertMessage(conversationId, userId, text, photoId, replyTo)

	if err != nil {
		http.Error(w, "Cannot save message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(mess)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		context.Logger.WithError(err).Error("Cannot save message")
		return
	}

}

func (rt *_router) postGroupMessage(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
	userId, err := strconv.Atoi(params.ByName("userId"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if !rt.db.IDExists(userId) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	replyToStr := r.URL.Query().Get("replyTo")

	var replyTo *int

	if replyToStr != "" {
		v, err := strconv.Atoi(replyToStr)
		if err != nil {
			http.Error(w, "Invalid replyTo", http.StatusBadRequest)
			return
		}
		replyTo = &v
	}
	groupId, err := strconv.Atoi(params.ByName("groupId"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")
	isThere, err := rt.db.UserGroup(userId, groupId, timestamp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !isThere {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	contentType := r.Header.Get("Content-Type")
	isMultipart := strings.HasPrefix(contentType, "multipart/")

	const maxSize = 10 << 20 // 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	var text string
	var photoId *int

	// -------- CASE 1: JSON SENZA FOTO --------
	if !isMultipart {
		var payload struct {
			Body string `json:"bodyText"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		text = payload.Body
	} else {
		// -------- CASE 2: MULTIPART (TESTO + FACOLTATIVA FOTO) --------
		if err := r.ParseMultipartForm(maxSize); err != nil {
			http.Error(w, "Invalid multipart form", http.StatusBadRequest)
			return
		}

		text = r.FormValue("bodyText")

		file, header, err := r.FormFile("photo")

		if err == nil {

			mime := header.Header.Get("Content-Type")

			if mime != JPEG && mime != PNG {
				http.Error(w, "Invalid image type", http.StatusBadRequest)
				return
			}

			ext := ".png"
			if mime == JPEG {
				ext = ".jpg"
			}

			defer func(file multipart.File) {
				err := file.Close()
				if err != nil {
					context.Logger.WithError(err).Error("closing file")
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
			}(file)

			path := fmt.Sprintf("uploads/groups/%d_%d_%s%s",
				groupId, userId, context.ReqUUID, ext,
			)

			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				http.Error(w, "Cannot create directory", http.StatusInternalServerError)
				return
			}

			out, err := os.Create(path)
			if err != nil {
				http.Error(w, "Cannot save image", http.StatusInternalServerError)
				context.Logger.WithError(err).Error("Cannot save image")
				return
			}
			defer func(out *os.File) {
				err := out.Close()
				if err != nil {
					http.Error(w, "Cannot save image", http.StatusInternalServerError)
					context.Logger.Errorf("Cannot save image: %s", err)
				}
			}(out)

			if _, err := io.Copy(out, file); err != nil {
				http.Error(w, "Write failed", http.StatusInternalServerError)
				return
			}
			imgFile, err := os.Open(path)
			if err != nil {
				context.Logger.WithError(err).WithField("path", path).Error("Cannot open file")
				http.Error(w, "Cannot save image", http.StatusInternalServerError)
				return
			}

			img, _, err := image.Decode(imgFile)

			if err != nil {
				context.Logger.WithError(err).Error("Error decoding photo file")
				http.Error(w, "Cannot decode image", http.StatusInternalServerError)
				return
			}
			width := img.Bounds().Dx()
			height := img.Bounds().Dy()
			id, err := rt.db.InsertPhoto(path, width, height, mime)
			if err != nil {
				http.Error(w, "Cannot save photo to DB", http.StatusInternalServerError)
				return
			}

			photoId = &id
		}
	}

	// ALMENO TESTO O FOTO
	if text == "" && photoId == nil {
		http.Error(w, "Message must have text or photo", http.StatusBadRequest)
		return
	}
	mess, err := rt.db.InsertGroupMessage(groupId, userId, text, photoId, replyTo)
	if err != nil {
		context.Logger.WithError(err).Error("Error inserting message")
		http.Error(w, "Cannot save message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(mess)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		context.Logger.WithError(err).Error("Cannot save message")
		return
	}

}
