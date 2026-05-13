package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type Photo struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Mime   string `json:"mime"`
}

type Body struct {
	Text  string `json:"text"`
	Photo *Photo `json:"photo,omitempty"`
}

type Message struct {
	MessageId   int     `json:"messageId"`
	Body        Body    `json:"body"`
	Read        *string `json:"read,omitempty"`
	Time        string  `json:"time"`
	Sender      User    `json:"sender"`
	IsForwarded bool    `json:"isForwarded"`
	ReplyTo     *int    `json:"replyTo,omitempty"`
}

func (rt *_router) getMessages(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
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
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	messages, err := rt.db.GetMessages(conversationId, userId)
	if err != nil {
		context.Logger.WithError(err).Error("Error getting messages")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	apiMessages := make([]Message, 0, len(messages))
	for _, dbMsg := range messages {
		// 1. Gestione del Body
		var apiPhoto *Photo
		if dbMsg.Body.Photo != nil { // Se c'è una foto nel DB
			apiPhoto = &Photo{
				Url:    dbMsg.Body.Photo.Url,
				Width:  dbMsg.Body.Photo.Width,
				Height: dbMsg.Body.Photo.Height,
				Mime:   dbMsg.Body.Photo.Mime,
			}
		}

		apiBody := Body{
			Text:  dbMsg.Body.Text, // Può essere stringa vuota se è solo foto
			Photo: apiPhoto,        // Nil se è solo testo
		}

		// 2. Aggiungi il messaggio alla lista
		apiMessages = append(apiMessages, Message{
			MessageId:   dbMsg.MessageId,
			Body:        apiBody,    // Assegna la struct Body complessa
			Read:        dbMsg.Read, // Assumo sia *string anche nel DB
			Time:        dbMsg.Time, // O converti in stringa come preferisci
			Sender:      User{UserId: dbMsg.Sender.UserID, Username: dbMsg.Sender.Username},
			IsForwarded: dbMsg.IsForwarded,
			ReplyTo:     dbMsg.ReplyTo,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(MessagesResponse{Messages: apiMessages})
	if err != nil {
		context.Logger.WithError(err).Error("error encoding conversations")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}

func (rt *_router) getGroupMessages(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
	userId, err := strconv.Atoi(params.ByName("userId"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if !rt.db.IDExists(userId) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	groupId, err := strconv.Atoi(params.ByName("groupId"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")
	isThere, err := rt.db.UserGroup(userId, groupId, timestamp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !isThere {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	messages, err := rt.db.GetGroupMessages(groupId, userId)
	if err != nil {
		context.Logger.WithError(err).Error("Error getting messages")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	apiMessages := make([]Message, 0, len(messages))
	for _, dbMsg := range messages {
		// 1. Gestione del Body
		var apiPhoto *Photo
		if dbMsg.Body.Photo != nil { // Se c'è una foto nel DB
			apiPhoto = &Photo{
				Url:    dbMsg.Body.Photo.Url,
				Width:  dbMsg.Body.Photo.Width,
				Height: dbMsg.Body.Photo.Height,
				Mime:   dbMsg.Body.Photo.Mime,
			}
		}

		apiBody := Body{
			Text:  dbMsg.Body.Text, // Può essere stringa vuota se è solo foto
			Photo: apiPhoto,        // Nil se è solo testo
		}

		// 2. Aggiungi il messaggio alla lista
		apiMessages = append(apiMessages, Message{
			MessageId:   dbMsg.MessageId,
			Body:        apiBody,    // Assegna la struct Body complessa
			Read:        dbMsg.Read, // Assumo sia *string anche nel DB
			Time:        dbMsg.Time, // O converti in stringa come preferisci
			Sender:      User{UserId: dbMsg.Sender.UserID, Username: dbMsg.Sender.Username},
			IsForwarded: dbMsg.IsForwarded,
			ReplyTo:     dbMsg.ReplyTo,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(MessagesResponse{Messages: apiMessages})
	if err != nil {
		context.Logger.WithError(err).Error("error encoding conversations")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}
