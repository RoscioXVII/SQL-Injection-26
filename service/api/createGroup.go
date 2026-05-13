package api

import (
	"encoding/json"

	"net/http"

	"strconv"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type GroupRequest struct {
	Name         string   `json:"name"`
	Participants []string `json:"participants"`
}

func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
	userId, err := strconv.Atoi(params.ByName("userId"))
	if err != nil {
		context.Logger.WithError(err).Error("Error converting userId to int")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if !rt.db.IDExists(userId) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")

	var group GroupRequest
	err = json.NewDecoder(r.Body).Decode(&group)
	if err != nil {
		context.Logger.WithError(err).Error("Error decoding body")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	groupR, err := rt.db.CreateGroup(userId, group.Name, group.Participants, timestamp)

	if err != nil {
		context.Logger.WithError(err).Error("Error creating group")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(groupR)
	if err != nil {
		context.Logger.WithError(err).Error("Error encoding group")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}
