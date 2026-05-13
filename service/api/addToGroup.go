package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type RequestAddToGroup struct {
	Name string `json:"name"`
}

func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
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

	var requestAddToGroup RequestAddToGroup
	err = json.NewDecoder(r.Body).Decode(&requestAddToGroup)
	if err != nil {
		context.Logger.WithError(err).Error("Error in getGroup")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	userIdToAddToGroup, err := rt.db.SearchUserByUsername(requestAddToGroup.Name, timestamp)
	if err != nil {
		context.Logger.WithError(err).Error("Error in getGroup")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	err = rt.db.AddToGroup(groupId, userIdToAddToGroup)
	if err != nil {
		context.Logger.WithError(err).Error("Error in addGroup add partecipant")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	group, err := rt.db.GetGroup(groupId)
	if err != nil {
		context.Logger.WithError(err).Error("Error in getGroup Information")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(group)
	if err != nil {
		context.Logger.WithError(err).Error("Error in getGroup")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}
