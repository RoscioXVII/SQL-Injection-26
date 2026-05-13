package database

func (db *appdbimpl) UserByToken(token string) (int, error) {
	var userId int
	var username string

	err := db.c.QueryRow(`
		SELECT u.userId, uu.username
		FROM User u
		JOIN Login l ON u.userId = l.userId
		LEFT JOIN UserUsername uu ON uu.userId = u.userId
		WHERE l.loginId = ? AND l.Time <= CURRENT_TIMESTAMP
		ORDER BY l.time DESC
		LIMIT 1
	`, token).Scan(&userId, &username)

	if err != nil {
		return 0, err
	}

	return userId, nil
}
