package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type BodyHandle struct {
	Name string `json:"name"`
}

func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
	var body BodyHandle
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		context.Logger.WithError(err).Error("error decoding body")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			context.Logger.WithError(err).Error("error closing body")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}(r.Body)
	userId_creator, err := strconv.Atoi(params.ByName("userId"))
	if err != nil {
		context.Logger.WithError(err).Error("error parsing userId")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")
	conversation, err := rt.db.CreateConversation(userId_creator, body.Name, timestamp)
	if err != nil {
		context.Logger.WithError(err).Error("error creating conversation")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(conversation); err != nil {
		context.Logger.WithError(err).Error("error encoding conversation")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}
