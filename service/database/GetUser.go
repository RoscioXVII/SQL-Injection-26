package database

import (
	"database/sql"
	"errors"
)

type DBUser struct {
	UserId   int
	Username string
	Avatar   Avatar
}

func (db *appdbimpl) Users() ([]DBUser, error) {

	rows, err := db.c.Query(`
        SELECT
            u.userId,
            uu.username,
            p.URL,
            p.mime,
            p.width,
            p.height
        FROM User u
        JOIN UserUsername uu
          ON uu.userId = u.userId
         AND uu.updateId = (
             SELECT MAX(updateId)
             FROM UserUsername
             WHERE userId = u.userId
         )
        JOIN UsPhoto up
          ON up.userId = u.userId
         AND up.updateId = (
             SELECT MAX(updateId)
             FROM UsPhoto
             WHERE userId = u.userId
         )
        JOIN Photo p
          ON p.photoId = up.photoId
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DBUser

	for rows.Next() {
		var u DBUser
		if err := rows.Scan(
			&u.UserId,
			&u.Username,
			&u.Avatar.Url,
			&u.Avatar.Mime,
			&u.Avatar.Width,
			&u.Avatar.Height,
		); err != nil {
			return nil, err
		}
		result = append(result, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (db *appdbimpl) GetUserById(userId int) ([]DBUser, error) {
	var u DBUser

	err := db.c.QueryRow(`
        SELECT
            u.userId,
            uu.username,
            p.URL,
            p.mime,
            p.width,
            p.height
        FROM User u
        JOIN UserUsername uu
          ON uu.userId = u.userId
         AND uu.updateId = (
             SELECT MAX(updateId)
             FROM UserUsername
             WHERE userId = u.userId
         )
        JOIN UsPhoto up
          ON up.userId = u.userId
         AND up.updateId = (
             SELECT MAX(updateId)
             FROM UsPhoto
             WHERE userId = u.userId
         )
        JOIN Photo p
          ON p.photoId = up.photoId
        WHERE u.userId = ?
    `, userId).Scan(
		&u.UserId,
		&u.Username,
		&u.Avatar.Url,
		&u.Avatar.Mime,
		&u.Avatar.Width,
		&u.Avatar.Height,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []DBUser{}, errors.New("user not found")
		}
		return []DBUser{}, err
	}
	list := []DBUser{}
	list = append(list, u)
	return list, nil
}

func (db *appdbimpl) IDExists(userId int) bool {
	var tmp int
	err := db.c.QueryRow(`
		SELECT 1
		FROM User
		WHERE userId = ?
	`, userId).Scan(&tmp)

	return err == nil
}

func (db *appdbimpl) SearchUserByUsername(username string, time string) (int, error) {
	var id int
	err := db.c.QueryRow(`
SELECT uu.userId
FROM UserUsername uu
WHERE uu.username = ?
  AND uu.time <= ?
ORDER BY uu.time DESC
LIMIT 1`, username, time).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errors.New("user not found")
	}
	return id, nil
}

func (db *appdbimpl) GetUsername(userId int, time string) (string, error) {
	var id string
	err := db.c.QueryRow(`
SELECT uu.username
FROM UserUsername uu
WHERE uu.userId= ?
  AND uu.time <= ?
ORDER BY uu.time DESC
LIMIT 1`, userId, time).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("user not found")
	}
	return id, nil
}

func (db *appdbimpl) GroupExists(groupId int) bool {
	var tmp int
	err := db.c.QueryRow(`
		SELECT 1
		FROM Group_
		WHERE groupId = ?
	`, groupId).Scan(&tmp)

	return err == nil
}
