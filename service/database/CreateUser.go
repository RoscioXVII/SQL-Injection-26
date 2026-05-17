/*
Realized by @luckignolo32 (GitHub) - MIT License
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

/*
CreateUser registers a new account identified by the provided username and password.
If the username is already taken it returns the sentinel error "username già esistente".
On success it persists a default avatar, opens a new Login row and returns the freshly
issued userId, authentication token and login timestamp so the caller can transparently
sign the user in.
*/
func (db *appdbimpl) CreateUser(username string, password string) (string, string, time.Time, error) {
	var existingUserId int
	err := db.c.QueryRow(
		`SELECT u.userId
		 FROM User u
		 JOIN UserUsername uu ON u.userId = uu.userId
		 WHERE uu.username = ?
		 ORDER BY uu.updateId DESC
		 LIMIT 1`, username).Scan(&existingUserId)
	if err == nil {
		return "", "", time.Time{}, fmt.Errorf("username già esistente")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", "", time.Time{}, fmt.Errorf("failed to check username: %w", err)
	}

	res, err := db.c.Exec("INSERT INTO User (password) VALUES (?);", password)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to create user: %w", err)
	}

	newUserId, err := res.LastInsertId()
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to get new userId: %w", err)
	}

	_, err = db.c.Exec(
		"INSERT INTO UserUsername(userId, username) VALUES (?, ?);",
		newUserId, username)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to insert username: %w", err)
	}

	userId := int(newUserId)

	err = db.SetMyPhoto("assets/default/default-avatar-profile-icon-social-600nw-1906669723.png", 600, 600, "image/png", userId)
	if err != nil {
		return "", "", time.Time{}, err
	}

	token, err := uuid.NewV4()
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate token: %w", err)
	}

	_, err = db.c.Exec("INSERT INTO Login(userId, loginId) VALUES (?, ?)", userId, token.String())
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to insert login: %w", err)
	}

	var loginTime time.Time
	err = db.c.QueryRow("SELECT time FROM Login WHERE loginId = ?", token.String()).Scan(&loginTime)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to query login time: %w", err)
	}

	return fmt.Sprintf("%d", userId), token.String(), loginTime, nil
}
