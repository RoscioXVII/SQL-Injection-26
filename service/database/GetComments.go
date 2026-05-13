package database

import (
	"database/sql"
	"strconv"
	"time"
)

type Comment struct {
	CommentId int       `json:"commentId"`
	Emoji     string    `json:"emoji"`
	Time      time.Time `json:"time"`
	Sender    User      `json:"sender"`
}

func (db *appdbimpl) GetComments(messageId int) ([]Comment, error) {
	rows, err := db.c.Query(`
        SELECT
            c.commentId,
            c.emoji,
            c.time,
            c.userId,
            uu.username
        FROM Comment c
        JOIN UserUsername uu
            ON uu.userId = c.userId
            AND uu.updateId = (
                SELECT MAX(updateId)
                FROM UserUsername
                WHERE userId = c.userId
            )
        WHERE c.messageId = ?
        ORDER BY c.time DESC
    `, messageId)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			panic(err)
		}
	}(rows)

	comments := make([]Comment, 0)
	for rows.Next() {
		var c Comment
		var userId int
		var username string

		if err := rows.Scan(&c.CommentId, &c.Emoji, &c.Time, &userId, &username); err != nil {
			return nil, err
		}
		userIdS := strconv.Itoa(userId)

		c.Sender = User{
			UserID:   userIdS,
			Username: username,
		}
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}
