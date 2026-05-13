package database

import (
	"database/sql"
	"errors"
)

func (db *appdbimpl) GetPhoto(userId int, time string) (Avatar, error) {
	var avatar Avatar
	err := db.c.QueryRow(`
		SELECT p.URL, p.mime, p.width, p.height
		FROM Photo p
		JOIN UsPhoto up ON p.photoId = up.photoId
		WHERE up.userId = ? AND up.time <= ?
		ORDER BY up.time DESC
		LIMIT 1
	`, userId, time).Scan(&avatar.Url, &avatar.Mime, &avatar.Width, &avatar.Height)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Avatar{}, errors.New("no photo")
		}
		return Avatar{}, err
	}

	return avatar, nil
}
