package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

type User struct {
	Username string `json:"username"`
	UserID   string `json:"userId"`
}

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

// Message rappresenta un messaggio o un inoltro
type Message struct {
	MessageId   int     `json:"messageId"`
	Body        Body    `json:"body"`
	Read        *string `json:"read,omitempty"`
	Time        string  `json:"time"`
	Sender      User    `json:"sender"`
	IsForwarded bool    `json:"isForwarded"`
	ReplyTo     *int    `json:"replyTo,omitempty"`
}

const READ = "read"
const RECEIVED = "received"

func (db *appdbimpl) GetMessages(conversationId int, viewerId int) ([]Message, error) {
	rows, err := db.c.Query(`
		SELECT DISTINCT
			m.messageId,
			m.time,
			m.sender,
			m.replyTo,
			uu.username,
			CASE WHEN m.originalMessage IS NOT NULL THEN 1 ELSE 0 END AS isForwarded
		FROM Message m
		JOIN UserUsername uu
  			ON uu.userId = m.sender
			 AND uu.updateId = (
				 SELECT MAX(updateId)
				 FROM UserUsername
				 WHERE userId = m.sender
			 )
		WHERE m.conversationId = ?
		  AND NOT EXISTS (
		      SELECT 1
		      FROM DeletedMessage d
		      WHERE d.messageId = m.messageId
		  )
		ORDER BY m.time DESC
	`, conversationId)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)

	messages := make([]Message, 0)

	for rows.Next() {
		var messageId int
		var time string
		var senderId int
		var senderUsername string
		var isForwarded int
		var replyTo *int

		if err := rows.Scan(&messageId, &time, &senderId, &replyTo, &senderUsername, &isForwarded); err != nil {
			return nil, err
		}

		// Ottieni i dati del messaggio originale solo per popolare Body
		originalMsg, err := db.OriginalMessageInfo(messageId)
		if err != nil {
			return nil, err
		}

		// Popola il messaggio corrente
		msg := Message{
			MessageId:   messageId,
			Time:        time,
			Sender:      User{UserID: fmt.Sprint(senderId), Username: senderUsername},
			Body:        originalMsg.Body, // include testo/foto dall'originale se inoltro
			IsForwarded: isForwarded == 1,
			ReplyTo:     replyTo,
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// ---- Gestione Read / Unread ----
	for i := range messages {
		senderId, err := strconv.Atoi(messages[i].Sender.UserID)
		if err != nil {
			return nil, err
		}

		if senderId != viewerId {
			// Messaggi ricevuti → segna come letti
			_, err = db.c.Exec(`
				INSERT OR IGNORE INTO ReadMessage(userId, messageId)
				VALUES (?, ?)
			`, viewerId, messages[i].MessageId)
			if err != nil {
				return nil, err
			}
			messages[i].Read = nil
		} else {
			// Messaggi inviati → stato read/unread
			var exists int
			err = db.c.QueryRow(`
				SELECT 1
				FROM ReadMessage
				WHERE userId != ? AND messageId = ?
			`, viewerId, messages[i].MessageId).Scan(&exists)

			switch {
			case errors.Is(err, sql.ErrNoRows):
				status := RECEIVED
				messages[i].Read = &status
			case err != nil:
				return nil, err
			default:
				status := READ
				messages[i].Read = &status
			}
		}
	}

	return messages, nil
}

func (db *appdbimpl) GetGroupMessages(groupId int, viewerId int) ([]Message, error) {
	rows, err := db.c.Query(`
		SELECT DISTINCT
			m.messageId,
			m.time,
			m.sender,
			m.replyTo,
			uu.username,
			CASE WHEN m.originalMessage IS NOT NULL THEN 1 ELSE 0 END AS isForwarded
		FROM Message m
		JOIN UserUsername uu
			ON uu.userId = m.sender
			AND uu.updateId = (
				SELECT MAX(updateId)
				FROM UserUsername
				WHERE userId = m.sender
			)
		WHERE m.groupId = ?
		  AND NOT EXISTS (
		      SELECT 1
		      FROM DeletedMessage d
		      WHERE d.messageId = m.messageId
		        AND d.userId = ?
		  )
		ORDER BY m.time DESC
	`, groupId, viewerId)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)

	messages := make([]Message, 0)

	for rows.Next() {
		var messageId int
		var time string
		var senderId int
		var senderUsername string
		var isForwarded int
		var replyTo *int

		if err := rows.Scan(&messageId, &time, &senderId, &replyTo, &senderUsername, &isForwarded); err != nil {
			return nil, err
		}

		// Ottieni il contenuto originale solo per Body
		originalMsg, err := db.OriginalMessageInfo(messageId)
		if err != nil {
			return nil, err
		}

		// Popola il messaggio corrente senza duplicare l’originale
		msg := Message{
			MessageId:   messageId,
			Time:        time,
			Sender:      User{UserID: fmt.Sprint(senderId), Username: senderUsername},
			Body:        originalMsg.Body, // testo/foto originale se inoltro
			IsForwarded: isForwarded == 1,
			ReplyTo:     replyTo,
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// ---- Gestione Read / Unread ----
	for i := range messages {
		senderId, err := strconv.Atoi(messages[i].Sender.UserID)
		if err != nil {
			return nil, err
		}

		if senderId != viewerId {
			// messaggi ricevuti → segna come letti
			_, err = db.c.Exec(`
            INSERT OR IGNORE INTO ReadMessage(userId, messageId)
            VALUES (?, ?)
        `, viewerId, messages[i].MessageId)
			if err != nil {
				return nil, err
			}
			continue
		}

		// Conta componenti attivi del gruppo
		var totalMembers int
		err = db.c.QueryRow(`
    SELECT COUNT(*)
    FROM Components
    WHERE groupId = ?
      AND timeLeft IS NULL
      AND userId != ?
`, groupId, senderId).Scan(&totalMembers)
		if err != nil {
			return nil, err
		}

		// Conta letture dei componenti attivi
		var readCount int
		err = db.c.QueryRow(`
			SELECT COUNT(DISTINCT rm.userId)
			FROM ReadMessage rm
			JOIN Components c ON c.userId = rm.userId
			WHERE rm.messageId = ?
			  AND c.groupId = ?
			  AND c.timeLeft IS NULL
		`, messages[i].MessageId, groupId).Scan(&readCount)
		if err != nil {
			return nil, err
		}

		if readCount == totalMembers {
			status := READ
			messages[i].Read = &status
		} else {
			status := RECEIVED
			messages[i].Read = &status
		}
	}

	return messages, nil
}
