package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type Comment struct {
	CommentId int       `json:"commentId"`
	Emoji     string    `json:"emoji"`
	Time      time.Time `json:"time"`
	Sender    User      `json:"sender"`
}

func (rt *_router) getComments(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
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

	comments, err := rt.db.GetComments(messageId)
	if err != nil {
		context.Logger.WithError(err).Error("Error getting comments")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	apiComments := make([]Comment, 0, len(comments))
	for _, dbC := range comments {
		apiComments = append(apiComments, Comment{
			CommentId: dbC.CommentId,
			Emoji:     dbC.Emoji,
			Time:      dbC.Time,
			Sender: User{
				UserId:   dbC.Sender.UserID,
				Username: dbC.Sender.Username,
			},
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(CommentsResponse{Comments: apiComments})
	if err != nil {
		context.Logger.WithError(err).Error("error encoding conversations")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}

func (rt *_router) getGroupComments(
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

	// --- get comments ---
	comments, err := rt.db.GetComments(messageId)
	if err != nil {
		context.Logger.WithError(err).Error("Error getting comments")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	apiComments := make([]Comment, 0, len(comments))
	for _, dbC := range comments {
		apiComments = append(apiComments, Comment{
			CommentId: dbC.CommentId,
			Emoji:     dbC.Emoji,
			Time:      dbC.Time,
			Sender: User{
				UserId:   dbC.Sender.UserID,
				Username: dbC.Sender.Username,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(CommentsResponse{Comments: apiComments}); err != nil {
		context.Logger.WithError(err).Error("Error encoding comments")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
