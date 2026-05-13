package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type ForwardMessageRequest struct {
	ForwardToConversation []int `json:"forwardToConversation"`
	ForwardToGroup        []int `json:"forwardToGroup"`
}

func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
	// --- userId ---
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

	originalMessageId := messageId

	// --- body ---
	var request ForwardMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		context.Logger.WithError(err).Error("Error decoding request")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// --- forward verso conversazioni ---
	for _, convId := range request.ForwardToConversation {
		isThere, err := rt.db.UserConversation(userId, convId)
		if err != nil {
			context.Logger.WithError(err).Error("Error checking user conversation")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if !isThere {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		err = rt.db.ForwardToConversation(userId, convId, originalMessageId)
		if err != nil {
			context.Logger.WithError(err).Error("Error forwarding to conversation")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")

	for _, groupId := range request.ForwardToGroup {
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

		err = rt.db.ForwardToGroup(userId, groupId, originalMessageId)
		if err != nil {
			context.Logger.WithError(err).Error("Error forwarding to group")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (rt *_router) forwardGroupMessage(
	w http.ResponseWriter,
	r *http.Request,
	params httprouter.Params,
	context reqcontext.RequestContext,
) {
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

	// --- Date header ---
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
	originalMessageId := messageId

	// --- body ---
	var request ForwardMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		context.Logger.WithError(err).Error("Error decoding request body")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// --- forward to conversations ---
	for _, convId := range request.ForwardToConversation {
		isThere, err := rt.db.UserConversation(userId, convId)
		if err != nil {
			context.Logger.WithError(err).Error("Error checking user conversation")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if !isThere {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		if err := rt.db.ForwardToConversation(userId, convId, originalMessageId); err != nil {
			context.Logger.WithError(err).Error("Error forwarding to conversation")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	// --- forward to groups ---
	for _, targetGroupId := range request.ForwardToGroup {
		isThere, err := rt.db.UserGroup(userId, targetGroupId, timestamp)
		if err != nil {
			context.Logger.WithError(err).Error("Error checking user group")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if !isThere {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		if err := rt.db.ForwardToGroup(userId, targetGroupId, originalMessageId); err != nil {
			context.Logger.WithError(err).Error("Error forwarding to group")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
