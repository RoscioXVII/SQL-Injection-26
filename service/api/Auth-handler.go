package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// authHandler è un middleware che controlla l'header Authorization e verifica il token nel DB SQLite3
func (rt *_router) authHandler(fn httpRouterHandler) httpRouterHandler {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.Logger.Error(errors.New("invalid auth header"))
			http.Error(w, "Unauthorized: missing or invalid token", http.StatusUnauthorized)
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

		idToken, err := rt.db.UserByToken(token)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.Logger.WithError(err).Warn("User by token failed")
				http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
				return
			}
			ctx.Logger.WithError(err).Warn("Internal server error")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		userId, err := strconv.Atoi(ps.ByName("userId"))
		if err != nil {
			ctx.Logger.WithError(err).Warn("User by token failed")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if userId != idToken {
			ctx.Logger.WithError(err).Warn("User by token failed")
			http.Error(w, "Unauthorized Action", http.StatusUnauthorized)
			return
		}

		fn(w, r, ps, ctx)
	}
}
