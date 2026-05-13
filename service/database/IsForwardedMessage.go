package database

import (
	"database/sql"
	"errors"
	"fmt"
)

func (db *appdbimpl) IsForwardedMessage(messageId int) (bool, error) {
	var exists int
	err := db.c.QueryRow(`
        SELECT 1
        FROM Message m
        WHERE m.originalMessage IS NOT NULL AND m.messageId = ?`, messageId).Scan(&exists)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (db *appdbimpl) OriginalMessageInfo(messageId int) (Message, error) {
	var msg Message
	var originalMessageId sql.NullInt64

	// Controlla se il messaggio è un inoltro
	err := db.c.QueryRow(`
        SELECT originalMessage
        FROM Message
        WHERE messageId = ?`, messageId).Scan(&originalMessageId)
	if err != nil {
		return msg, err
	}

	if originalMessageId.Valid {
		// Se è un inoltro, prendi ricorsivamente il messaggio originale
		originalMsg, err := db.OriginalMessageInfo(int(originalMessageId.Int64))
		if err != nil {
			return msg, err
		}
		// Imposta IsForwarded = true perché questo messaggio è un inoltro
		originalMsg.IsForwarded = true
		return originalMsg, nil
	}

	// Messaggio originale → prendi i dettagli
	var text sql.NullString
	var photoURL, photoMime sql.NullString
	var photoWidth, photoHeight sql.NullInt64
	var senderId int
	var senderUsername string

	err = db.c.QueryRow(`
        SELECT
            m.messageId,
            m.text,
            p.url,
            p.mime,
            p.width,
            p.height,
            m.time,
            uu.username,
            m.sender
        FROM Message m
        LEFT JOIN Photo p ON m.photoId = p.photoId
        JOIN UserUsername uu ON uu.userId = m.sender
        WHERE m.messageId = ?
        ORDER BY uu.updateId DESC
        LIMIT 1
    `, messageId).Scan(
		&msg.MessageId,
		&text,
		&photoURL,
		&photoMime,
		&photoWidth,
		&photoHeight,
		&msg.Time,
		&senderUsername,
		&senderId,
	)
	if err != nil {
		return msg, err
	}

	// Popola il corpo del messaggio
	if text.Valid {
		msg.Body.Text = text.String
	}
	if photoURL.Valid {
		msg.Body.Photo = &Photo{
			Url:    photoURL.String,
			Mime:   photoMime.String,
			Width:  int(photoWidth.Int64),
			Height: int(photoHeight.Int64),
		}
	}

	// Popola il mittente
	msg.Sender = User{
		UserID:   fmt.Sprint(senderId),
		Username: senderUsername,
	}

	// Messaggio originale → IsForwarded = false
	msg.IsForwarded = false

	return msg, nil
}
