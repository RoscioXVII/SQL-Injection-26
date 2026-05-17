/*
Realized by @luckignolo32 (GitHub) - MIT License
*/
package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	UserId string    `json:"userId"`
	Token  string    `json:"token"`
	Time   time.Time `json:"time"`
}

/*
postRegister implements the POST /register endpoint. It parses the JSON body, delegates
account creation to the database layer and, upon success, returns a freshly minted
session (userId + token + timestamp) so the frontend can immediately log the user in.
A duplicate username yields HTTP 409 with a structured JSON error payload; a malformed
or empty payload yields HTTP 400.
*/
func (rt *_router) postRegister(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "text/plain")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		context.Logger.WithError(err).Error("Error reading body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req RegisterRequest
	if err := json.Unmarshal(body, &req); err != nil {
		context.Logger.WithError(err).Error("Error parsing JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Username e password sono richiesti",
		})
		return
	}

	var resp RegisterResponse
	resp.UserId, resp.Token, resp.Time, err = rt.db.CreateUser(req.Name, req.Password)
	if err != nil {
		if err.Error() == "username già esistente" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Username già esistente",
			})
			return
		}

		context.Logger.WithError(err).Error("Error creating user")
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		context.Logger.WithError(err).Error("Error encoding response")
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}
}
