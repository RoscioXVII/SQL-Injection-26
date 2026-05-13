package database

import (
	"errors"
)

// CreateConversation crea una conversazione tra userId e username.
// Se la conversazione esiste già, restituisce i dati della conversazione esistente.
func (db *appdbimpl) CreateConversation(userId int, username string, time string) (Conversation, error) {
	componentB, err := db.SearchUserByUsername(username, time)
	if err != nil {
		return Conversation{}, err
	}

	res, err := db.c.Exec(`INSERT INTO Conversation (component_A, component_B) VALUES (?, ?)`, userId, componentB)
	if err != nil {
		// intercetta correttamente l'errore del trigger SQLite
		var sqliteErr interface{ Error() string }
		if errors.As(err, &sqliteErr) && sqliteErr.Error() == "Conversation already created" {
			// la conversazione esiste già, recupera i dati
			var conv Conversation
			row := db.c.QueryRow(`
                SELECT conversationId
                FROM Conversation
                WHERE (component_A = ? AND component_B = ?)
                   OR (component_A = ? AND component_B = ?)
            `, userId, componentB, componentB, userId)

			var id int
			if err := row.Scan(&id); err != nil {
				return Conversation{}, err
			}

			conv.ConversationID = id
			conv.Name = username
			conv.Avatar, err = db.GetPhoto(componentB, time)
			if err != nil {
				return Conversation{}, err
			}
			return conv, nil
		}
		return Conversation{}, err
	}

	// conversazione appena creata
	var conversation Conversation
	id, err := res.LastInsertId()
	if err != nil {
		return Conversation{}, err
	}
	conversation.ConversationID = int(id)
	conversation.Name = username
	conversation.Avatar, err = db.GetPhoto(componentB, time)
	if err != nil {
		return Conversation{}, err
	}

	return conversation, nil
}
