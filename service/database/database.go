/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"time"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	GetName() (string, error)
	SetName(name string) error
	CreateSession(string) (string, string, time.Time, error)
	UserByToken(token string) (int, error)
	Users() ([]DBUser, error)
	GetUserById(userId int) ([]DBUser, error)
	SetMyUserName(int, string) error
	SetMyPhoto(url string, width int, height int, mime string, userId int) error
	IDExists(token int) bool
	SetGroupPhoto(url string, width int, height int, mime string, groupId int) error
	GetConversations(userId int) ([]Conversation, error)
	CreateConversation(userId int, username string, time string) (Conversation, error)
	SearchUserByUsername(username string, time string) (int, error)
	GetPhoto(userId int, time string) (Avatar, error)
	GetConversationById(currentUserId int, conversationId int) (Conversation, error)
	UserConversation(userId int, conversationId int) (bool, error)
	GetMessages(conversationId int, userId int) ([]Message, error)
	InsertPhoto(url string, width int, height int, mime string) (int, error)
	InsertMessage(conversationId int, userId int, text string, photoId *int, replyTo *int) (Message, error)
	GetUsername(userId int, time string) (string, error)
	IsRead(messageId int, userId int) (string, error)
	UserMessage(userId int, messageId int) (bool, error)
	DeleteMessage(userId int, messageId int) error
	DeleteForwardedMessage(userId int, forwardedId int) error
	UserGroup(userId int, groupId int, date string) (bool, error)
	ForwardToConversation(userId int, conversationId int, messageId int) error
	ForwardToGroup(userId int, groupId int, messageId int) error
	IsForwardedMessage(messageId int) (bool, error)
	GetOriginalMessageId(messageId int) (int, error)
	ForwardToConversationWithParent(userId, conversationId, originalMsgId, parentFwdId int) error
	ForwardToGroupWithParent(userId, groupId, originalMsgId, parentFwdId int) error
	OriginalMessageInfo(messageId int) (Message, error)
	GetComments(messageId int) ([]Comment, error)
	PostComment(messageId int, userId int, emoji string) error
	MessageComment(messageId int, commentId int) (bool, error)
	CommentUser(commentId int, userId int) (bool, error)
	UnComment(commentId int) error
	GetGroups(userId int) ([]Group, error)
	CreateGroup(userId int, name string, partecipants []string, time string) (Group, error)
	GetGroup(groupId int) (Group, error)
	SetGroupName(groupId int, name string) error
	AddToGroup(groupId int, userId int) error
	LeaveGroup(groupId int, userId int) error
	InsertGroupMessage(groupId int, userId int, text string, photoId *int, replyTo *int) (Message, error)
	GetGroupMessages(groupId int, viewerId int) ([]Message, error)
	GroupExists(groupId int) bool
	DeleteGroupMessage(messageId int) error
	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	impl := &appdbimpl{c: db}

	if err := impl.initSchema(); err != nil {
		return nil, err
	}

	return impl, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
