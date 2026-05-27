package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Collegamento al database SQLite
	db, err := sql.Open("sqlite3", "./data/decaf.db")
	if err != nil {
		log.Fatal("Errore apertura DB:", err)
	}
	defer db.Close()

	fmt.Println("Inizio migrazione automatica delle password...")

	// 1. Recuperiamo tutti gli utenti e le loro password attuali
	query := `
		SELECT u.userId, uu.username, u.password 
		FROM User u
		JOIN UserUsername uu ON u.userId = uu.userId
		WHERE uu.updateId = (
			SELECT MAX(updateId) FROM UserUsername WHERE userId = u.userId
		)
	`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Errore durante la lettura degli utenti:", err)
	}
	defer rows.Close()

	aggiornati := 0
	saltati := 0

	for rows.Next() {
		var userId int
		var username, currentPassword string

		if err := rows.Scan(&userId, &username, &currentPassword); err != nil {
			log.Println("Errore lettura riga:", err)
			continue
		}

		// 2. Controlliamo se la password è GIÀ un hash bcrypt valido.

		_, err = bcrypt.Cost([]byte(currentPassword)) // restituisce errore se la stringa non è un vero hash.
		if err == nil {
			fmt.Printf("  L'utente %s ha già un hash valido, salto...\n", username)
			saltati++
			continue
		}

		// 3. Generiamo il VERO hash bcrypt partendo dalla password attualmente salvata
		hashed, err := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf(" Errore generazione hash per %s: %v\n", username, err)
			continue
		}

		// 4. Aggiorniamo il database sovrascrivendo la vecchia password con l'hash
		updateQuery := "UPDATE User SET password = ? WHERE userId = ?"
		_, err = db.Exec(updateQuery, string(hashed), userId)
		if err != nil {
			log.Printf(" Errore salvataggio nel DB per %s: %v\n", username, err)
		} else {
			fmt.Printf(" Hash generato per: %s (Per loggarti usa l'esatta stringa che aveva prima: %s)\n", username, currentPassword)
			aggiornati++
		}
	}

	fmt.Printf("\nMigrazione completata! Utenti aggiornati: %d, Utenti saltati: %d\n", aggiornati, saltati)
}
