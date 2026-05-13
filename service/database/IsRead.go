package database

import (
	"database/sql"
	"errors"
)

func (db *appdbimpl) IsRead(messageId int, userId int) (string, error) {
	var dummy int
	err := db.c.QueryRow(`
		SELECT 1
		FROM ReadMessage
		WHERE messageId = ? AND userId = ?
	`, messageId, userId).Scan(&dummy)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "received", nil
		}
		return "", err
	}
	return "read", nil
}
