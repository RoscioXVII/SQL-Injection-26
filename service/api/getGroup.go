package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getGroup(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
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
	groupInfo, err := rt.db.GetGroup(groupId)
	if err != nil {
		context.Logger.WithError(err).Error("Error in getGroup")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(groupInfo)
	if err != nil {
		context.Logger.WithError(err).Error("Error encoding group")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}
