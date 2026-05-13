package database

func (db *appdbimpl) InsertPhoto(url string, width int, height int, mime string) (int, error) {
	res, err := db.c.Exec(`
INSERT INTO Photo (url, width, height, mime) VALUES (?,?,?,?)
`, url, width, height, mime)
	if err != nil {
		return 0, err
	}
	photoId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(photoId), nil
}
