package database

import (
	"database/sql"
	"strconv"
)

type Group struct {
	GroupId      int      `json:"groupId"`
	Name         string   `json:"name"`
	Photo        Avatar   `json:"upload"`
	Participants []User   `json:"participants"`
	LastMessage  *Message `json:"lastMessage,omitempty"`
}

func (db *appdbimpl) GetGroups(userId int) ([]Group, error) {
	groups := make([]Group, 0)

	// prendo tutti i gruppi in cui l'utente è membro attivo
	rows, err := db.c.Query(`
		SELECT groupId, name
		FROM Group_
		WHERE groupId IN (
			SELECT groupId
			FROM Components
			WHERE userId = ? AND timeLeft IS NULL
		)
	`, userId)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)

	for rows.Next() {
		var g Group
		err := rows.Scan(&g.GroupId, &g.Name)
		if err != nil {
			return nil, err
		}

		// Avatar del gruppo
		var avatar Avatar
		err = db.c.QueryRow(`
			SELECT p.url, p.mime, p.width, p.height
			FROM GroupPhoto gp
			JOIN Photo p ON p.photoId = gp.photoId
			WHERE gp.groupId = ?
			ORDER BY gp.updateId DESC
			LIMIT 1
		`, g.GroupId).Scan(
			&avatar.Url, &avatar.Mime, &avatar.Width, &avatar.Height,
		)
		if err != nil {
			return nil, err
		} else {
			g.Photo = avatar
		}

		// Partecipanti
		urows, err := db.c.Query(`
			SELECT u.userId, uu.username
			FROM Components c
			JOIN User u ON u.userId = c.userId
			JOIN UserUsername uu ON uu.userId = u.userId
			WHERE c.groupId = ? AND c.timeLeft IS NULL
			  AND uu.updateId = (
				  SELECT MAX(updateId)
				  FROM UserUsername
				  WHERE userId = u.userId
			  )
		`, g.GroupId)
		if err != nil {
			return nil, err
		}
		var participants []User
		for urows.Next() {
			var u User
			if err := urows.Scan(&u.UserID, &u.Username); err != nil {
				err := urows.Close()
				if err != nil {
					return nil, err
				}
				return nil, err
			}
			participants = append(participants, u)
		}
		if err := urows.Err(); err != nil {
			return nil, err
		}
		err = urows.Close()
		if err != nil {
			return nil, err
		}
		g.Participants = participants

		// Ultimo messaggio del gruppo
		var lastMsgID sql.NullInt64
		var lastMsgText sql.NullString
		var lastMsgTime sql.NullString
		var lastMsgSenderID sql.NullInt64
		var lastMsgOriginal sql.NullInt64
		var lastMsgSenderUsername sql.NullString
		var msgPhotoURL sql.NullString
		var msgPhotoWidth sql.NullInt64
		var msgPhotoHeight sql.NullInt64
		var msgPhotoMime sql.NullString

		err = db.c.QueryRow(`
    SELECT
        m.messageId,
        COALESCE(om.text, m.text)        AS text,
        m.time,
        COALESCE(om.sender, m.sender)   AS sender,
        m.originalMessage,
        uu.username,
        p.url,
        p.width,
        p.height,
        p.mime
    FROM Message m
    LEFT JOIN Message om ON om.messageId = m.originalMessage
    LEFT JOIN UserUsername uu ON uu.userId = COALESCE(om.sender, m.sender)
        AND uu.updateId = (
            SELECT MAX(updateId)
            FROM UserUsername
            WHERE userId = COALESCE(om.sender, m.sender)
        )
    LEFT JOIN Photo p ON p.photoId = COALESCE(om.photoId, m.photoId)
    WHERE m.groupId = ?
    ORDER BY m.time DESC
    LIMIT 1
`, g.GroupId).Scan(
			&lastMsgID, &lastMsgText, &lastMsgTime, &lastMsgSenderID,
			&lastMsgOriginal, &lastMsgSenderUsername,
			&msgPhotoURL, &msgPhotoWidth, &msgPhotoHeight, &msgPhotoMime,
		)

		if err == nil && lastMsgID.Valid {
			var photo *Photo
			if msgPhotoURL.Valid {
				photo = &Photo{
					Url:    msgPhotoURL.String,
					Width:  int(msgPhotoWidth.Int64),
					Height: int(msgPhotoHeight.Int64),
					Mime:   msgPhotoMime.String,
				}
			}

			g.LastMessage = &Message{
				MessageId: int(lastMsgID.Int64),
				Body: Body{
					Text:  lastMsgText.String,
					Photo: photo,
				},
				Time: lastMsgTime.String,
				Sender: User{
					UserID:   strconv.Itoa(int(lastMsgSenderID.Int64)),
					Username: lastMsgSenderUsername.String,
				},
				IsForwarded: lastMsgOriginal.Valid,
			}
		} else {
			g.LastMessage = nil
		}

		groups = append(groups, g)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}
