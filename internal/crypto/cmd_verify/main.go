package main

import (
	"database/sql"
	"log"
	"wecom-gateway/internal/crypto"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	encKey := crypto.GenerateKeyFromPassphrase("default-change-me")
	log.Printf("encKey len=%d", len(encKey))

	db, err := sql.Open("sqlite3", "../../../data/wecom.db")
	if err != nil { log.Fatal(err) }
	defer db.Close()

	rows, err := db.Query("SELECT name, corp_name, agent_id, secret_enc FROM wecom_apps")
	if err != nil { log.Fatal(err) }
	defer rows.Close()

	for rows.Next() {
		var name, corpName, secretEnc string
		var agentID int64
		rows.Scan(&name, &corpName, &agentID, &secretEnc)

		decrypted, err := crypto.DecryptString(secretEnc, encKey)
		if err != nil {
			log.Printf("❌ App %s (corp=%s agent_id=%d): decrypt FAILED: %v", name, corpName, agentID, err)
		} else {
			log.Printf("✅ App %s (corp=%s agent_id=%d): decrypted secret = %s", name, corpName, agentID, decrypted)
		}
	}
}
