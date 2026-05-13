package database

func (db *appdbimpl) LeaveGroup(groupId int, userId int) error {
	_, err := db.c.Exec(`UPDATE Components
SET timeLeft = CURRENT_TIMESTAMP
WHERE groupId = ?
  AND userId = ?
  AND timeLeft IS NULL;`, groupId, userId)
	if err != nil {
		return err
	}
	return nil
}
