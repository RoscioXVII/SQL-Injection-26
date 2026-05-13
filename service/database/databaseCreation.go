package database

import (
	"fmt"
)

const SQLschema = `
	PRAGMA journal_mode=WAL;
	PRAGMA busy_timeout = 5000;
	PRAGMA foreign_keys = ON;


---GROUP---
CREATE TABLE IF NOT EXISTS Group_(
    groupId INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    creator INTEGER   REFERENCES User(userId) NOT NULL
);

CREATE TABLE IF NOT EXISTS Components(
    updateId INTEGER PRIMARY KEY AUTOINCREMENT,
    groupId INTEGER   REFERENCES Group_(groupId) NOT NULL,
    userId INTEGER   REFERENCES User(userId) NOT NULL,
    timeAdded TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    timeLeft TIMESTAMP
);

---USER---
CREATE TABLE IF NOT EXISTS User(
    userId INTEGER PRIMARY KEY AUTOINCREMENT,
    password VARCHAR(40) NOT NULL
);

CREATE TABLE IF NOT EXISTS Login(
	loginId VARCHAR(36) PRIMARY KEY,
	userId INTEGER REFERENCES User(userId) NOT NULL,
	time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS UserUsername(
    updateId INTEGER PRIMARY KEY AUTOINCREMENT,
    time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    userId INTEGER   REFERENCES User(userId) NOT NULL,
    username TEXT NOT NULL
);
---PHOTO---
CREATE TABLE IF NOT EXISTS Photo (
    photoId INTEGER PRIMARY KEY AUTOINCREMENT,
    URL TEXT NOT NULL,
    mime TEXT NOT NULL
        CHECK (mime LIKE '%/%'),
    width INTEGER NOT NULL CHECK (width > 0),
    height INTEGER NOT NULL CHECK ( height > 0 )
);

CREATE TABLE IF NOT EXISTS UsPhoto(
    photoId INTEGER   REFERENCES Photo(photoId) NOT NULL,
    userId INTEGER   REFERENCES User(userId) NOT NULL,
    time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updateId INTEGER PRIMARY KEY AUTOINCREMENT
);

CREATE TABLE IF NOT EXISTS GroupPhoto(
    updateId INTEGER PRIMARY KEY AUTOINCREMENT,
    time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    photoId INTEGER   REFERENCES Photo(photoId) NOT NULL,
    groupId INTEGER   REFERENCES Group_ (groupId) NOT NULL

);


---CONVERSATIONS---
CREATE TABLE IF NOT EXISTS Conversation(
    conversationId INTEGER PRIMARY KEY AUTOINCREMENT,
    component_A INTEGER   REFERENCES User(userId) NOT NULL,
    component_B INTEGER   REFERENCES User(userId) NOT NULL
);

---MESSAGES---
CREATE TABLE IF NOT EXISTS Message(
    messageId INTEGER PRIMARY KEY AUTOINCREMENT,
    conversationId INTEGER   REFERENCES Conversation(conversationId),
    groupId INTEGER   REFERENCES Group_(groupId),
    text TEXT,
    photoId INTEGER   REFERENCES Photo(photoId),
    time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sender INTEGER   REFERENCES User(userId),
	originalMessage INTEGER   REFERENCES Message(messageId),
	replyTo INTEGER   REFERENCES Message(messageId),
    CHECK ( (originalMessage IS NULL AND (text IS NOT NULL OR photoId IS NOT NULL)) OR
(originalMessage IS NOT NULL AND (text IS NULL AND photoId IS NULL) ) ),
    CHECK (
        (conversationId IS NOT NULL AND groupId IS NULL)
            OR
        (conversationId IS NULL AND groupId IS NOT NULL)
        )
);



CREATE TABLE IF NOT EXISTS ReadMessage(
    userId INTEGER   REFERENCES User(userId) NOT NULL,
    messageId INTEGER   REFERENCES Message(messageId) NOT NULL,
    time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (userId,messageId)
);


CREATE TABLE IF NOT EXISTS DeletedMessage(
	userId 	  	INTEGER   REFERENCES User(userId) NOT NULL,
	messageId 	INTEGER   REFERENCES Message(messageId),
	time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (userId,messageId)
);

---COMMENT---

CREATE TABLE IF NOT EXISTS Comment(
    commentId INTEGER PRIMARY KEY AUTOINCREMENT,
    time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    emoji TEXT NOT NULL,
    userId INTEGER   REFERENCES User(userId) NOT NULL,
    messageId INTEGER   REFERENCES Message(messageId) NOT NULL,
	UNIQUE(messageId,userId)
);



---TRIGGER---

CREATE TRIGGER IF NOT EXISTS CHECK_USERNAME
    BEFORE INSERT ON UserUsername
    FOR EACH ROW
BEGIN
    SELECT
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM UserUsername uu
                WHERE uu.username = NEW.username
                  AND uu.userId <> NEW.userId
                  AND uu.updateId = (
                    SELECT MAX(updateId)
                    FROM UserUsername
                    WHERE userId = uu.userId
                )
            )
            THEN RAISE(ABORT, 'Username already in use')
            END;
END;

CREATE TRIGGER IF NOT EXISTS CHECK_CONVERSATION
    BEFORE INSERT ON Conversation
    FOR EACH ROW
BEGIN
    SELECT
        CASE
            WHEN EXISTS(
                SELECT 1
                FROM Conversation c
                WHERE (c.component_A == NEW.component_A AND c.component_B == NEW.component_B)
                   OR (c.component_A == NEW.component_B AND c.component_B == NEW.component_A)
            )
            THEN RAISE(ABORT,'Conversation already created')
            END;
END;

CREATE TRIGGER IF NOT EXISTS CHECK_GROUP_COMPONENTS
    BEFORE INSERT ON Components
    FOR EACH ROW
BEGIN
    SELECT
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM Components c
                WHERE c.groupId = NEW.groupId
                  AND c.userId = NEW.userId
                  AND c.timeLeft IS NULL
            )
                THEN RAISE(ABORT, 'User already in this group')
            END;
END;


CREATE TRIGGER IF NOT EXISTS CHECK_MESSAGE_TARGET
BEFORE INSERT ON Message
FOR EACH ROW
BEGIN
    SELECT
        CASE
            -- Messaggio / inoltro in CONVERSAZIONE
            WHEN NEW.conversationId IS NOT NULL
                 AND NOT EXISTS (
                     SELECT 1
                     FROM Conversation c
                     WHERE c.conversationId = NEW.conversationId
                       AND (c.component_A = NEW.sender
                            OR c.component_B = NEW.sender)
                 )
            THEN RAISE(ABORT, 'Sender not part of the conversation')

            -- Messaggio / inoltro in GRUPPO
            WHEN NEW.groupId IS NOT NULL
                 AND NOT EXISTS (
                     SELECT 1
                     FROM Components comp
                     WHERE comp.groupId = NEW.groupId
                       AND comp.userId = NEW.sender
                       AND comp.timeLeft IS NULL
                 )
            THEN RAISE(ABORT, 'Sender not member of the group')
        END;
END;

CREATE TRIGGER IF NOT EXISTS CHECK_ORIGINAL_MESSAGE_VISIBILITY
BEFORE INSERT ON Message
FOR EACH ROW
WHEN NEW.originalMessage IS NOT NULL
BEGIN
    SELECT
        CASE
            -- Originale in CONVERSAZIONE
            WHEN EXISTS (
                SELECT 1
                FROM Message m
                WHERE m.messageId = NEW.originalMessage
                  AND m.conversationId IS NOT NULL
            )
            AND NOT EXISTS (
                SELECT 1
                FROM Message m
                JOIN Conversation c
                    ON c.conversationId = m.conversationId
                WHERE m.messageId = NEW.originalMessage
                  AND (c.component_A = NEW.sender
                       OR c.component_B = NEW.sender)
            )
            THEN RAISE(ABORT,
                'Forwarder cannot see original message (conversation)')

            -- Originale in GRUPPO
            WHEN EXISTS (
                SELECT 1
                FROM Message m
                WHERE m.messageId = NEW.originalMessage
                  AND m.groupId IS NOT NULL
            )
            AND NOT EXISTS (
                SELECT 1
                FROM Message m
                JOIN Components comp
                    ON comp.groupId = m.groupId
                WHERE m.messageId = NEW.originalMessage
                  AND comp.userId = NEW.sender
                  AND comp.timeLeft IS NULL
            )
            THEN RAISE(ABORT,
                'Forwarder cannot see original message (group)')
        END;
END;

`

func (db *appdbimpl) initSchema() error {
	_, err := db.c.Exec(SQLschema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	return nil
}
