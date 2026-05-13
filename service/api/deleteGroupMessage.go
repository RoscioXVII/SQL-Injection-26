package api

import (
	"net/http"
	"strconv"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) deleteGroupMessage(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
	userId, err := strconv.Atoi(params.ByName("userId"))
	if err != nil {
		context.Logger.WithError(err).Error("Error converting userId to int")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if !rt.db.IDExists(userId) {
		context.Logger.WithError(err).Error("User does not exist")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	groupId, err := strconv.Atoi(params.ByName("groupId"))
	if err != nil {
		context.Logger.WithError(err).Error("Error converting groupId to int")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")

	isThere, err := rt.db.UserGroup(userId, groupId, timestamp)
	if err != nil {
		context.Logger.WithError(err).Error("Error getting user group")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !isThere {
		context.Logger.WithError(err).Error("User group not found")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	messageId, err := strconv.Atoi(params.ByName("messageId"))
	if err != nil {
		context.Logger.WithError(err).Error("Error converting messageId to int")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	isThere, err = rt.db.UserMessage(userId, messageId)
	if err != nil {
		context.Logger.WithError(err).Error("Error getting user message")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !isThere {
		context.Logger.WithError(err).Error("User message not found")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	err = rt.db.DeleteGroupMessage(messageId)
	if err != nil {
		context.Logger.WithError(err).Error("Error deleting message")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
