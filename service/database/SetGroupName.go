package database

func (db *appdbimpl) SetGroupName(groupId int, name string) error {

	_, err := db.c.Exec(`UPDATE Group_ SET name = ? WHERE groupId = ?`, name, groupId)
	if err != nil {
		return err
	}

	return nil
}
