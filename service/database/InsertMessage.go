package database

import "database/sql"

func (db *appdbimpl) InsertMessage(conversationId int, userId int, text string, photoId *int, replyTo *int) (Message, error) {

	var res sql.Result
	var err error

	var replyToVal sql.NullInt64
	if replyTo != nil {
		replyToVal = sql.NullInt64{
			Int64: int64(*replyTo),
			Valid: true,
		}
	} else {
		replyToVal = sql.NullInt64{Valid: false}
	}

	if photoId != nil {
		res, err = db.c.Exec(
			`INSERT INTO Message (conversationId, sender, text, photoId, replyTo)
			 VALUES (?, ?, ?, ?,?)`,
			conversationId, userId, text, *photoId, replyToVal,
		)
	} else {
		res, err = db.c.Exec(
			`INSERT INTO Message (conversationId, sender, text, replyTo)
			 VALUES (?, ?, ?, ?)`,
			conversationId, userId, text, replyToVal,
		)
	}

	if err != nil {
		return Message{}, err
	}

	messageId, err := res.LastInsertId()
	if err != nil {
		return Message{}, err
	}

	// ritorno il mess appena creato
	var msg Message

	var photoUrl sql.NullString
	var photoWidth, photoHeight sql.NullInt64
	var photoMime sql.NullString

	err = db.c.QueryRow(`
		SELECT
			m.messageId,
			m.text,
			m.replyTo,

			p.URL,
			p.width,
			p.height,
			p.mime,

			m.time,

			u.userId,
			uu.username,

			m.originalMessage IS NOT NULL

		FROM Message m

		JOIN User u
			ON u.userId = m.sender

		JOIN UserUsername uu
			ON uu.userId = u.userId
		   AND uu.updateId = (
				SELECT MAX(updateId)
				FROM UserUsername
				WHERE userId = u.userId
		   )

		LEFT JOIN Photo p
			ON p.photoId = m.photoId

		WHERE m.messageId = ?
	`, messageId).Scan(
		&msg.MessageId,
		&msg.Body.Text,
		&msg.ReplyTo,

		&photoUrl,
		&photoWidth,
		&photoHeight,
		&photoMime,

		&msg.Time,

		&msg.Sender.UserID,
		&msg.Sender.Username,

		&msg.IsForwarded,
	)

	if err != nil {
		return Message{}, err
	}

	// se c'è la foto la metto
	if photoUrl.Valid {
		msg.Body.Photo = &Photo{
			Url:    photoUrl.String,
			Width:  int(photoWidth.Int64),
			Height: int(photoHeight.Int64),
			Mime:   photoMime.String,
		}
	}

	return msg, nil
}
func (db *appdbimpl) InsertGroupMessage(groupId int, userId int, text string, photoId *int, replyTo *int) (Message, error) {

	var res sql.Result
	var err error

	var replyToVal sql.NullInt64
	if replyTo != nil {
		replyToVal = sql.NullInt64{
			Int64: int64(*replyTo),
			Valid: true,
		}
	} else {
		replyToVal = sql.NullInt64{Valid: false}
	}

	if photoId != nil {
		res, err = db.c.Exec(
			`INSERT INTO Message (groupId, sender, text, photoId, replyTo)
			 VALUES (?, ?, ?, ?,?)`,
			groupId, userId, text, *photoId, replyToVal,
		)
	} else {
		res, err = db.c.Exec(
			`INSERT INTO Message (groupId, sender, text, replyTo)
			 VALUES (?, ?, ?,?)`,
			groupId, userId, text, replyToVal,
		)
	}

	if err != nil {
		return Message{}, err
	}

	messageId, err := res.LastInsertId()
	if err != nil {
		return Message{}, err
	}

	var msg Message

	var photoUrl sql.NullString
	var photoWidth, photoHeight sql.NullInt64
	var photoMime sql.NullString

	err = db.c.QueryRow(`
		SELECT
			m.messageId,
			m.text,
			m.replyTo,

			p.URL,
			p.width,
			p.height,
			p.mime,

			m.time,

			u.userId,
			uu.username,

			m.originalMessage IS NOT NULL

		FROM Message m

		JOIN User u
			ON u.userId = m.sender

		JOIN UserUsername uu
			ON uu.userId = u.userId
		   AND uu.updateId = (
				SELECT MAX(updateId)
				FROM UserUsername
				WHERE userId = u.userId
		   )

		LEFT JOIN Photo p
			ON p.photoId = m.photoId

		WHERE m.messageId = ?
	`, messageId).Scan(
		&msg.MessageId,
		&msg.Body.Text,
		&msg.ReplyTo,

		&photoUrl,
		&photoWidth,
		&photoHeight,
		&photoMime,

		&msg.Time,

		&msg.Sender.UserID,
		&msg.Sender.Username,

		&msg.IsForwarded,
	)

	if err != nil {
		return Message{}, err
	}

	if photoUrl.Valid {
		msg.Body.Photo = &Photo{
			Url:    photoUrl.String,
			Width:  int(photoWidth.Int64),
			Height: int(photoHeight.Int64),
			Mime:   photoMime.String,
		}
	}

	return msg, nil
}
