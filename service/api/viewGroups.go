package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type Group struct {
	GroupId      int      `json:"groupId"`
	Name         string   `json:"name"`
	Photo        Upload   `json:"upload"`
	Participants []User   `json:"participants"`
	LastMessage  *Message `json:"lastMessage,omitempty"`
}

func (rt *_router) viewGroups(w http.ResponseWriter, r *http.Request, params httprouter.Params, context reqcontext.RequestContext) {
	userId, err := strconv.Atoi(params.ByName("userId"))
	if err != nil {
		context.Logger.WithError(err).Error("Error converting userId to int")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	groups, err := rt.db.GetGroups(userId)
	if err != nil {
		context.Logger.WithError(err).Error("Error getting groups")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	apiGroups := make([]Group, 0, len(groups))
	for _, dbGroup := range groups {
		var apiParticipants []User
		for _, dbUser := range dbGroup.Participants {
			apiParticipants = append(apiParticipants, User{
				UserId:   dbUser.UserID,
				Username: dbUser.Username,
			})
		}
		var lastMsg *Message
		if dbGroup.LastMessage != nil {
			var photo *Photo
			if dbGroup.LastMessage.Body.Photo != nil {
				photo = &Photo{
					Url:    dbGroup.LastMessage.Body.Photo.Url,
					Mime:   dbGroup.LastMessage.Body.Photo.Mime,
					Width:  dbGroup.LastMessage.Body.Photo.Width,
					Height: dbGroup.LastMessage.Body.Photo.Height,
				}
			}
			lastMsg = &Message{
				MessageId: dbGroup.LastMessage.MessageId,
				Body: Body{
					Text:  dbGroup.LastMessage.Body.Text,
					Photo: photo,
				},
				Time: dbGroup.LastMessage.Time,
				Sender: User{
					UserId:   dbGroup.LastMessage.Sender.UserID,
					Username: dbGroup.LastMessage.Sender.Username,
				},
				IsForwarded: dbGroup.LastMessage.IsForwarded,
			}
		}
		apiGroups = append(apiGroups, Group{
			GroupId: dbGroup.GroupId,
			Name:    dbGroup.Name,
			Photo: Upload{
				Url:    dbGroup.Photo.Url,
				Mime:   dbGroup.Photo.Mime,
				Width:  dbGroup.Photo.Width,
				Height: dbGroup.Photo.Height,
			},
			Participants: apiParticipants,
			LastMessage:  lastMsg,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(GroupsResponse{Groups: apiGroups})
	if err != nil {
		context.Logger.WithError(err).Error("Error encoding group")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}
