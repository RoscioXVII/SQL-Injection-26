package database

import "database/sql"

// GetGroup restituisce tutte le informazioni di un gruppo dato il suo ID
func (db *appdbimpl) GetGroup(groupId int) (Group, error) {
	var group Group

	// Recupera l'ID e il nome del gruppo
	row := db.c.QueryRow(`SELECT groupId, name FROM Group_ WHERE groupId = ?`, groupId)
	err := row.Scan(&group.GroupId, &group.Name)
	if err != nil {
		return group, err
	}

	// Recupera la foto del gruppo
	row = db.c.QueryRow(`
		SELECT p.URL, p.mime, p.width, p.height
		FROM Photo p
		JOIN GroupPhoto gp ON gp.photoId = p.photoId
		WHERE gp.groupId = ?
		ORDER BY gp.time DESC
		LIMIT 1
	`, groupId)

	var photo Avatar
	err = row.Scan(&photo.Url, &photo.Mime, &photo.Width, &photo.Height)
	if err != nil {
		// Se non c'è foto, possiamo lasciarla vuota senza error
		photo = Avatar{}
	}
	group.Photo = photo

	// Recupera i partecipanti del gruppo
	rows, err := db.c.Query(`
		SELECT u.userId, uu.username
		FROM Components c
		JOIN User u ON u.userId = c.userId
		JOIN UserUsername uu ON uu.userId = u.userId
		WHERE c.groupId = ? AND c.timeLeft IS NULL
		  AND uu.updateId = (
				SELECT MAX(updateId)
				FROM UserUsername
				WHERE userId = u.userId
		  )
	`, groupId)
	if err != nil {
		return group, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			panic(err)
		}
	}(rows)

	var participants []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.UserID, &user.Username); err != nil {
			return group, err
		}
		participants = append(participants, user)
	}
	group.Participants = participants
	if err := rows.Err(); err != nil {
		return Group{}, err
	}
	return group, nil
}
