package database

func (db *appdbimpl) PostComment(messageId int, userId int, emoji string) error {
	_, err := db.c.Exec(`INSERT INTO Comment(messageId, userId, emoji) VALUES (?, ?, ?)`, messageId, userId, emoji)
	if err != nil {
		return err
	}
	return nil
}
