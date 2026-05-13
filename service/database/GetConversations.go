package database

import (
	"database/sql"
	"strconv"
)

type Avatar struct {
	Url    string `json:"url"`
	Mime   string `json:"mime"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Conversation struct {
	ConversationID int      `json:"conversationId"`
	Name           string   `json:"name"`
	Avatar         Avatar   `json:"avatar"`
	LastMessage    *Message `json:"lastMessage,omitempty"`
}

func (db *appdbimpl) GetConversations(userId int) ([]Conversation, error) {
	rows, err := db.c.Query(`SELECT
    c.conversationId,
    uu.username,
    p.url,
    p.mime,
    p.width,
    p.height,
    m.messageId,
    COALESCE(om.text, m.text)              AS text, -- se om.text è NULL inserisci m.text
    m.time,
    COALESCE(om.sender, m.sender)          AS sender, -- se om.sender è NULL inserisci m.sender
    m.originalMessage,
    COALESCE(om.photoId, m.photoId)        AS photoId, -- se om.photoId è NULL inserisci m.photoId
    uuMsg.username,
    pp.url,
    pp.width,
    pp.height,
    pp.mime
FROM Conversation c
JOIN UserUsername uu ON uu.userId = CASE
    WHEN c.component_A = ? THEN c.component_B
    ELSE c.component_A
END
AND uu.updateId = (
    SELECT MAX(updateId)
    FROM UserUsername
    WHERE userId = uu.userId
)
LEFT JOIN UsPhoto up ON up.userId = uu.userId
AND up.updateId = (
    SELECT up2.updateId
    FROM UsPhoto up2
    WHERE up2.userId = uu.userId
    ORDER BY up2.updateId DESC
    LIMIT 1
)
LEFT JOIN Photo p ON p.photoId = up.photoId
LEFT JOIN Message m ON m.conversationId = c.conversationId
  AND m.time = (
      SELECT MAX(time)
      FROM Message
      WHERE conversationId = c.conversationId
  )

-- self join sul messaggio originale
LEFT JOIN Message om ON om.messageId = m.originalMessage
LEFT JOIN UserUsername uuMsg ON uuMsg.userId = m.sender
  AND uuMsg.updateId = (
      SELECT MAX(updateId)
      FROM UserUsername
      WHERE userId = m.sender
  )
LEFT JOIN Photo pp ON pp.photoId = m.photoId
WHERE c.component_A = ? OR c.component_B = ?;`, userId, userId, userId)

	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)

	conversations := make([]Conversation, 0)

	for rows.Next() {
		var conv Conversation

		// variabili ultimo messaggio
		var lastMsgID sql.NullInt64
		var lastMsgText sql.NullString
		var lastMsgTime sql.NullString
		var lastMsgSenderID sql.NullInt64
		var lastMsgOriginal sql.NullInt64
		var lastMsgPhotoID sql.NullInt64
		var lastMsgSenderUsername sql.NullString
		var msgPhotoURL sql.NullString
		var msgPhotoWidth sql.NullInt64
		var msgPhotoHeight sql.NullInt64
		var msgPhotoMime sql.NullString

		err := rows.Scan(
			&conv.ConversationID,
			&conv.Name,
			&conv.Avatar.Url,
			&conv.Avatar.Mime,
			&conv.Avatar.Width,
			&conv.Avatar.Height,
			&lastMsgID,
			&lastMsgText,
			&lastMsgTime,
			&lastMsgSenderID,
			&lastMsgOriginal,
			&lastMsgPhotoID,
			&lastMsgSenderUsername,
			&msgPhotoURL,
			&msgPhotoWidth,
			&msgPhotoHeight,
			&msgPhotoMime,
		)
		if err != nil {
			return nil, err
		}

		if lastMsgID.Valid {
			var photo *Photo
			if msgPhotoURL.Valid {
				photo = &Photo{
					Url:    msgPhotoURL.String,
					Width:  int(msgPhotoWidth.Int64),
					Height: int(msgPhotoHeight.Int64),
					Mime:   msgPhotoMime.String,
				}
			}
			conv.LastMessage = &Message{
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
			conv.LastMessage = nil
		}

		conversations = append(conversations, conv)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}
