package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// SessionRequest : Request's schema
type SessionRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type SessionResponse struct {
	UserId string    `json:"userId"`
	Token  string    `json:"token"`
	Time   time.Time `json:"time"`
}

func (rt *_router) postSession(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "text/plain")

	// Reading body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		context.Logger.WithError(err).Error("Error reading body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parsing JSON
	var req SessionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		context.Logger.WithError(err).Error("Error parsing JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var resp SessionResponse
	// Execute query and return response
	resp.UserId, resp.Token, resp.Time, err = rt.db.CreateSession(req.Name) //mettere anche pw dopo il name

	if err != nil {
		context.Logger.WithError(err).Error("Error creating session")
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		context.Logger.WithError(err).Error("Error encoding response")
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}
}
