package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type setUsernameRequest struct {
	NewUsername string `json:"newusername"`
}

func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
	autHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(autHeader, "Bearer ")
	token = strings.TrimSpace(token)
	if token == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	userIdStr := params.ByName("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		context.Logger.WithError(err).Error("Invalid user id")
		return
	}
	var req setUsernameRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		context.Logger.WithError(err).Error("Invalid request")
		return
	}

	usernamePropost := req.NewUsername
	if usernamePropost == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		context.Logger.WithError(err).Warn("username must not be empty")
		return
	}
	if !rt.db.IDExists(userId) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		context.Logger.WithError(err).Warn("username does not exist")
		return
	}

	err = rt.db.SetMyUserName(userId, usernamePropost)
	if err != nil {
		if err.Error() == "username already in use" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			err = json.NewEncoder(w).Encode(map[string]string{
				"error": "username already in use",
			})
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				context.Logger.WithError(err).Warn("failed to encode user response")
				return
			}
			return
		}
		context.Logger.WithError(err).Error("db error while setting username")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var u User
	u.UserId = strconv.Itoa(userId)
	u.Username = usernamePropost

	err = json.NewEncoder(w).Encode(u)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		context.Logger.WithError(err).Warn("failed to encode user response")
		return

	}

}
