package database

import (
	"database/sql"
	"errors"
)

func (db *appdbimpl) UserMessage(userId int, messageId int) (bool, error) {
	var exists int
	err := db.c.QueryRow(`
        SELECT 1
        FROM Message m
        WHERE m.sender = ? AND m.messageId = ?`, userId, messageId).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
