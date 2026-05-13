package database

import (
	"database/sql"
	"errors"
)

func (db *appdbimpl) UserConversation(userId int, conversationId int) (bool, error) {
	var tmp int
	err := db.c.QueryRow(`
        SELECT 1
        FROM Conversation c
        WHERE c.conversationId = ?
          AND (c.component_A = ? OR c.component_B = ?)
    `, conversationId, userId, userId).Scan(&tmp)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (db *appdbimpl) UserGroup(userId int, groupId int, date string) (bool, error) {
	var tmp int
	err := db.c.QueryRow(`
        SELECT 1
        FROM Components c
        WHERE c.groupId = ?
          AND c.userId = ? AND c.timeAdded <= ? AND (c.timeLeft IS NULL OR c.timeLeft >= ?)
    `, groupId, userId, date, date).Scan(&tmp)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
