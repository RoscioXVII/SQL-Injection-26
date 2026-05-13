package database

import (
	"errors"
)

func (db *appdbimpl) SetMyPhoto(url string, width int, height int, mime string, userId int) error {
	res, err := db.c.Exec(`
INSERT INTO Photo (url, width, height, mime) VALUES (?,?,?,?)
`, url, width, height, mime)
	if err != nil {
		return err
	}
	photoId, err := res.LastInsertId()
	if err != nil {
		return err
	}
	if !db.IDExists(userId) {
		var err404 = errors.New("User not found")
		return err404
	}
	_, err = db.c.Exec(`
	INSERT INTO UsPhoto (photoId,userId) VALUES (?,?)
`, int(photoId), userId)
	if err != nil {
		return err
	}
	return nil
}

func (db *appdbimpl) SetGroupPhoto(url string, width int, height int, mime string, groupId int) error {
	res, err := db.c.Exec(`
INSERT INTO Photo (url, width, height, mime) VALUES (?,?,?,?)
`, url, width, height, mime)
	if err != nil {
		return err
	}
	photoId, err := res.LastInsertId()
	if err != nil {
		return err
	}
	if !db.GroupExists(groupId) {
		var err404 = errors.New("group not found")
		return err404
	}
	_, err = db.c.Exec(`
	INSERT INTO GroupPhoto (photoId,groupId) VALUES (?,?)
`, int(photoId), groupId)
	if err != nil {
		return err
	}
	return nil
}
