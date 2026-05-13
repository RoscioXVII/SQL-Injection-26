package database

func (db *appdbimpl) ForwardToConversationWithParent(userId, conversationId, originalMsgId, parentFwdId int) error {
	_, err := db.c.Exec(`
        INSERT INTO ForwardedMessage (originalMessage, parentForwardedId, targetConv, targetGroup, forwarder)
        VALUES (?, ?, ?, NULL, ?)`,
		originalMsgId, parentFwdId, conversationId, userId)
	return err
}

func (db *appdbimpl) ForwardToGroupWithParent(userId, groupId, originalMsgId, parentFwdId int) error {
	_, err := db.c.Exec(`
        INSERT INTO ForwardedMessage (originalMessage, parentForwardedId, targetConv, targetGroup, forwarder)
        VALUES (?, ?, NULL, ?, ?)`,
		originalMsgId, parentFwdId, groupId, userId)
	return err
}
