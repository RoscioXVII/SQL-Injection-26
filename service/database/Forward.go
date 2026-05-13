package database

func (db *appdbimpl) ForwardToConversation(userId int, conversationId int, originalMessageId int) error {
	_, err := db.c.Exec(
		`INSERT INTO Message(conversationId, sender, originalMessage)
		 VALUES (?,?,?)`,
		conversationId, userId, originalMessageId,
	)
	return err
}

func (db *appdbimpl) ForwardToGroup(userId int, groupId int, originalMessageId int) error {
	_, err := db.c.Exec(
		`INSERT INTO Message(groupId, sender, originalMessage)
		 VALUES (?,?,?)`,
		groupId, userId, originalMessageId,
	)
	return err
}
