package api

import (
	"encoding/json"
	"net/http"
	"time"

	"strconv"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type CommentRequest struct {
	Emoji string `json:"emoji"`
}

func isEmoji(s string) bool {
	for _, r := range s {
		switch {
		case r >= 0x1F300 && r <= 0x1F5FF:
			return true
		case r >= 0x1F600 && r <= 0x1F64F:
			return true
		case r >= 0x1F680 && r <= 0x1F6FF:
			return true
		case r >= 0x2600 && r <= 0x26FF:
			return true
		case r >= 0x2700 && r <= 0x27BF:
			return true
		}
	}
	return false
}

func (rt *_router) postComment(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
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
	conversationId, err := strconv.Atoi(params.ByName("conversationId"))
	if err != nil {
		context.Logger.WithError(err).Error("Error converting conversationId to int")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	isThere, err := rt.db.UserConversation(userId, conversationId)
	if err != nil {
		context.Logger.WithError(err).Error("Error getting user conversation")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !isThere {
		context.Logger.WithError(err).Error("User conversation not found")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	messageId, err := strconv.Atoi(params.ByName("messageId"))
	if err != nil {
		context.Logger.WithError(err).Error("Error converting messageId to int")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	var req CommentRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		context.Logger.WithError(err).Error("Error decoding request body")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if req.Emoji == "" || !isEmoji(req.Emoji) {
		context.Logger.WithError(err).Error("No emoji specified or isn't a valid emoji")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	err = rt.db.PostComment(messageId, userId, req.Emoji)
	if err != nil {
		context.Logger.WithError(err).Error("Error posting comment")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	comments, err := rt.db.GetComments(messageId)
	if err != nil {
		context.Logger.WithError(err).Error("Error getting comments")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(comments)
	if err != nil {
		context.Logger.WithError(err).Error("error encoding conversations")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}

func (rt *_router) postGroupComment(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
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

	var req CommentRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		context.Logger.WithError(err).Error("Error decoding request body")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if req.Emoji == "" || !isEmoji(req.Emoji) {
		context.Logger.WithError(err).Error("No emoji specified or isn't a valid emoji")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	err = rt.db.PostComment(messageId, userId, req.Emoji)
	if err != nil {
		context.Logger.WithError(err).Error("Error posting comment")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	comments, err := rt.db.GetComments(messageId)
	if err != nil {
		context.Logger.WithError(err).Error("Error getting comments")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(comments)
	if err != nil {
		context.Logger.WithError(err).Error("error encoding conversations")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}
