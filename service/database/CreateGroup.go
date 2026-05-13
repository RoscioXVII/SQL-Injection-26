package database

import "strconv"

func (db *appdbimpl) CreateGroup(userId int, name string, partecipants []string, time string) (Group, error) {
	var group Group
	row, err := db.c.Exec(`INSERT INTO Group_(name,creator) VALUES (?,?)`, name, userId)
	if err != nil {
		return Group{}, err
	}
	groupId, err := row.LastInsertId()
	if err != nil {
		return Group{}, err
	}

	group.Name = name
	group.GroupId = int(groupId)
	group.Photo.Url = "assets/default/default-avatar-profile-icon-social-600nw-1906669723.png"
	group.Photo.Width = 600
	group.Photo.Height = 600
	group.Photo.Mime = "image/png"
	err = db.SetGroupPhoto(group.Photo.Url, group.Photo.Width, group.Photo.Height, group.Photo.Mime, int(groupId))

	if err != nil {
		return Group{}, err
	}

	group.Participants = make([]User, 0)
	for _, participant := range partecipants {
		var user User
		userIdParticipant, err := db.SearchUserByUsername(participant, time)
		if err != nil {
			return Group{}, err
		}
		_, err = db.c.Exec(`INSERT INTO Components(groupId,userId) VALUES (?,?)`, groupId, userIdParticipant)
		if err != nil {
			return Group{}, err
		}
		user.UserID = strconv.Itoa(userIdParticipant)
		user.Username = participant
		group.Participants = append(group.Participants, user)
	}
	_, err = db.c.Exec(`INSERT INTO Components(groupId,userId) VALUES (?,?)`, groupId, userId)
	if err != nil {
		return Group{}, err
	}
	var user User
	user.UserID = strconv.Itoa(userId)
	user.Username, err = db.GetUsername(userId, time)
	if err != nil {
		return Group{}, err
	}
	group.Participants = append(group.Participants, user)
	return group, nil

}
