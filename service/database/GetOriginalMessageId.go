package database

func (db *appdbimpl) GetOriginalMessageId(forwardedId int) (int, error) {
	var originalMsgId int
	err := db.c.QueryRow(`
        SELECT originalMessage
        FROM ForwardedMessage
        WHERE forwardedId = ?`, forwardedId).Scan(&originalMsgId)
	if err != nil {
		return 0, err
	}
	return originalMsgId, nil
}
