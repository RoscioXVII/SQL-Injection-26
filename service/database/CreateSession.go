package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

func (db *appdbimpl) CreateSession(username string, pw string, isLogin bool) (string, string, time.Time, error) {

	var userId int
	var password string

	query := fmt.Sprintf(`SELECT u.userId, u.password
              FROM User u
              JOIN UserUsername uu ON u.userId = uu.userId
              WHERE uu.username = '%s'
              ORDER BY uu.updateId DESC
              LIMIT 1`, username)

	err := db.c.QueryRow(query).Scan(&userId, &password)

	// ================= LOGIN =================
	if isLogin {

		query := fmt.Sprintf(`
		SELECT u.userId
		FROM User u
		JOIN UserUsername uu ON u.userId = uu.userId
		WHERE uu.username = '%s'
		AND u.password = '%s'
		LIMIT 1
	`, username, pw)

		err := db.c.QueryRow(query).Scan(&userId)

		if errors.Is(err, sql.ErrNoRows) {
			return "", "", time.Time{}, fmt.Errorf("credenziali errate")
		}

		if err != nil {
			return "", "", time.Time{}, err
		}
	} else {

		// ================= SIGNUP =================

		// Username già esistente
		if err == nil {
			return "", "", time.Time{}, fmt.Errorf("username già esistente")
		}

		// Errore vero del DB
		if !errors.Is(err, sql.ErrNoRows) {
			return "", "", time.Time{}, err
		}

		// Creazione utente
		res, err := db.c.Exec(
			"INSERT INTO User(password) VALUES (?);",
			pw,
		)
		if err != nil {
			return "", "", time.Time{}, fmt.Errorf("failed to create user: %w", err)
		}

		newUserId, err := res.LastInsertId()
		if err != nil {
			return "", "", time.Time{}, fmt.Errorf("failed to get new userId: %w", err)
		}

		_, err = db.c.Exec(
			"INSERT INTO UserUsername(userId, username) VALUES (?, ?);",
			newUserId,
			username,
		)
		if err != nil {
			return "", "", time.Time{}, fmt.Errorf("failed to insert username: %w", err)
		}

		userId = int(newUserId)

		err = db.SetMyPhoto(
			"assets/default/default-avatar-profile-icon-social-600nw-1906669723.png",
			600,
			600,
			"image/png",
			userId,
		)
		if err != nil {
			return "", "", time.Time{}, err
		}
	}

	// ================= SESSIONE =================

	token, err := uuid.NewV4()
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate token: %w", err)
	}

	_, err = db.c.Exec(
		"INSERT INTO Login(userId, loginId) VALUES (?, ?)",
		userId,
		token.String(),
	)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to insert login: %w", err)
	}

	var loginTime time.Time

	err = db.c.QueryRow(
		"SELECT time FROM Login WHERE loginId = ?",
		token.String(),
	).Scan(&loginTime)

	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to query login time: %w", err)
	}

	return fmt.Sprintf("%d", userId), token.String(), loginTime, nil
}
