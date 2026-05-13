package database

func (db *appdbimpl) DeleteMessage(userId int, messageId int) error {
	_, err := db.c.Exec(`INSERT OR IGNORE INTO DeletedMessage(userId,messageId) VALUES (?,?)`, userId, messageId)
	if err != nil {
		return err
	}
	return nil

}

func (db *appdbimpl) DeleteGroupMessage(messageId int) error {
	_, err := db.c.Exec(`
        INSERT OR IGNORE INTO DeletedMessage(userId, messageId)
        SELECT userId, ?
        FROM Components
        WHERE groupId = (SELECT groupId FROM Message WHERE messageId = ?)
    `, messageId, messageId)
	return err
}

func (db *appdbimpl) DeleteForwardedMessage(userId int, forwardedId int) error {
	_, err := db.c.Exec(`INSERT OR IGNORE INTO DeletedMessage(userId,forwardedId) VALUES (?,?)`, userId, forwardedId)
	if err != nil {
		return err
	}
	return nil

}
