package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type User struct {
	UserId   string `json:"userId"`
	Username string `json:"username"`
}

type Info struct {
	UserId        int    `json:"userId"`
	Username      string `json:"username"`
	AvatarProfile Upload `json:"avatar"`
}

func (rt *_router) searchUser(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	paramUserID := r.URL.Query().Get("userId")
	if paramUserID == "" {
		dbUsers, err := rt.db.Users()
		if err != nil {
			context.Logger.WithError(err).Error("Error converting userId to int")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return

		}
		var users []Info
		for _, u := range dbUsers {
			users = append(users, Info{
				UserId:   u.UserId,
				Username: u.Username,
				AvatarProfile: Upload{
					Url:    u.Avatar.Url,
					Mime:   u.Avatar.Mime,
					Width:  u.Avatar.Width,
					Height: u.Avatar.Height,
				},
			})
		}
		err = json.NewEncoder(w).Encode(UsersResponse{Users: users})
		if err != nil {
			context.Logger.WithError(err).Error("Error converting userId to int")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	} else {
		paramUserIDInt, err := strconv.Atoi(paramUserID)
		if err != nil {
			context.Logger.WithError(err).Error("Error converting userId to int")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		dbUsers, err := rt.db.GetUserById(paramUserIDInt)
		if err != nil {
			context.Logger.WithError(err).Error("Error converting userId to int")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		var users []Info
		for _, u := range dbUsers {
			users = append(users, Info{
				UserId:   u.UserId,
				Username: u.Username,
				AvatarProfile: Upload{
					Url:    u.Avatar.Url,
					Mime:   u.Avatar.Mime,
					Width:  u.Avatar.Width,
					Height: u.Avatar.Height,
				},
			})
		}
		err = json.NewEncoder(w).Encode(UsersResponse{Users: users})
		if err != nil {
			context.Logger.WithError(err).Error("Error converting userId to int")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}
}
