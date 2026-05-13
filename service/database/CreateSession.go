package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

func (db *appdbimpl) CreateSession(username string, pw string) (string, string, time.Time, error) {
	var userId int
	var password string
	// Controlla se l'utente esiste --- modificare qui + gestire pw sbagliata
	query := `SELECT u.userId, u.password
              FROM User u
              JOIN UserUsername uu ON u.userId = uu.userId
              WHERE uu.username = ?
              ORDER BY uu.updateId DESC
              LIMIT 1`

	err := db.c.QueryRow(query, username).Scan(&userId, &password)
	if err == nil { // utente esiste ma password errata
		if password != pw {
			return "", "", time.Time{}, fmt.Errorf("password errata")
		}
	}
	if errors.Is(err, sql.ErrNoRows) {
		// Se non esiste, crea l'utente
		res, err := db.c.Exec("INSERT INTO User (password) VALUES (?);", pw) // default + password
		if err != nil {
			return "", "", time.Time{}, fmt.Errorf("failed to create user: %w", err)
		}

		newUserId, err := res.LastInsertId()
		if err != nil {
			return "", "", time.Time{}, fmt.Errorf("failed to get new userId: %w", err)
		}

		// Inserisci il nuovo username
		_, err = db.c.Exec(
			"INSERT INTO UserUsername(userId, username) VALUES (?, ?);",
			newUserId, username)
		if err != nil {
			return "", "", time.Time{}, fmt.Errorf("failed to insert username: %w", err)
		}

		userId = int(newUserId)

		err = db.SetMyPhoto("assets/default/default-avatar-profile-icon-social-600nw-1906669723.png", 600, 600, "image/png", userId)
		if err != nil {
			return "", "", time.Time{}, err
		}
	}

	// Genera token UUID
	token, err := uuid.NewV4()
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate token: %w", err)
	}

	// Inserisci login con userId e token
	_, err = db.c.Exec("INSERT INTO Login(userId, loginId) VALUES (?, ?)", userId, token.String())
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to insert login: %w", err)
	}
	// Recupera il timestamp del login appena inserito
	var loginTime time.Time
	err = db.c.QueryRow("SELECT time FROM Login WHERE loginId = ?", token.String()).Scan(&loginTime)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to query login time: %w", err)
	}

	// Ritorna userId, token e timestamp
	return fmt.Sprintf("%d", userId), token.String(), loginTime, nil
}
