package database

import "errors"

func (db *appdbimpl) SetMyUserName(userId int, username string) error {
	_, err := db.c.Exec("INSERT INTO UserUsername (userId, username) VALUES (?, ?)", userId, username)
	if err != nil {
		return errors.New("username already in use")
	}
	return nil
}
