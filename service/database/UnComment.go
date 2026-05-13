package database

func (db *appdbimpl) UnComment(commentId int) error {
	_, err := db.c.Exec(`DELETE FROM Comment WHERE commentId=? `, commentId)
	if err != nil {
		return err
	}
	return nil
}
