package api

import (
	"net/http"
	"strconv"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) unComment(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
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

	// --- conversationId (origine) ---
	conversationId, err := strconv.Atoi(params.ByName("conversationId"))
	if err != nil {
		context.Logger.WithError(err).Error("Error converting conversationId")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	isThere, err := rt.db.UserConversation(userId, conversationId)
	if err != nil {
		context.Logger.WithError(err).Error("Error checking user conversation")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !isThere {
		context.Logger.WithError(err).Error("Error checking user conversation")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// --- messageId ---
	messageId, err := strconv.Atoi(params.ByName("messageId"))
	if err != nil {
		context.Logger.WithError(err).Error("Error converting messageId")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	commentId, err := strconv.Atoi(params.ByName("commentId"))
	if err != nil {
		context.Logger.WithError(err).Error("Error converting commentId")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	isThere, err = rt.db.MessageComment(messageId, commentId)
	if err != nil {
		context.Logger.WithError(err).Error("Error checking message comment")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !isThere {
		context.Logger.WithError(err).Error("Error checking message comment")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	isThere, err = rt.db.CommentUser(commentId, userId)
	if err != nil {
		context.Logger.WithError(err).Error("Error checking message comment")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !isThere {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	err = rt.db.UnComment(commentId)
	if err != nil {
		context.Logger.WithError(err).Error("Error unComment")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}

func (rt *_router) unGroupComment(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
	// --- userId ---
	userId, err := strconv.Atoi(params.ByName("userId"))
	if err != nil {
		context.Logger.WithError(err).Error("Error converting userId")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if !rt.db.IDExists(userId) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// --- groupId (origine) ---
	groupId, err := strconv.Atoi(params.ByName("groupId"))
	if err != nil {
		context.Logger.WithError(err).Error("Error converting groupId")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")

	// --- user must belong to origin group ---
	isThere, err := rt.db.UserGroup(userId, groupId, timestamp)
	if err != nil {
		context.Logger.WithError(err).Error("Error checking user group")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !isThere {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// --- messageId ---
	messageId, err := strconv.Atoi(params.ByName("messageId"))
	if err != nil {
		context.Logger.WithError(err).Error("Error converting messageId")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	commentId, err := strconv.Atoi(params.ByName("commentId"))
	if err != nil {
		context.Logger.WithError(err).Error("Error converting commentId")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	isThere, err = rt.db.MessageComment(messageId, commentId)
	if err != nil {
		context.Logger.WithError(err).Error("Error checking message comment")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !isThere {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	isThere, err = rt.db.CommentUser(commentId, userId)
	if err != nil {
		context.Logger.WithError(err).Error("Error checking message comment")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !isThere {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	err = rt.db.UnComment(commentId)
	if err != nil {
		context.Logger.WithError(err).Error("Error unComment")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
