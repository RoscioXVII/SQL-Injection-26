package database

func (db *appdbimpl) AddToGroup(groupId int, userId int) error {
	_, err := db.c.Exec(`INSERT INTO Components(groupId,userId) VALUES(?,?)`, groupId, userId)
	if err != nil {
		return err
	}
	return nil
}
