package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/pressly/goose/v3"
)

func Migrate(db *sql.DB) {
	// Zet de goose migratie directory
	goose.SetBaseFS(nil) // gebruik lokale filesysteem
	migrationsDir := "./db/migrations"

	// Controleer huidige versie
	current, err := goose.GetDBVersion(db)
	if err != nil {
		log.Fatalf("error getting DB version: %v", err)
	}
	fmt.Println("Current DB version:", current)

	// Voer alle nog niet toegepaste migraties uit
	if err := goose.Up(db, migrationsDir); err != nil {
		log.Fatalf("goose up: %v", err)
	}

	fmt.Println("✅ Migraties succesvol uitgevoerd!")
}

func StoreTelegram(db *sql.DB, deviceID string, tele Telegram, prevGasTimestamp time.Time) error {
	electricityJSON, _ := json.Marshal(tele.Electricity)
	gasJSON, _ := json.Marshal(tele.Gas)

	for _, entry := range []struct {
		Type string
		Data []byte
	}{
		{"electricity", electricityJSON},
		{"gas", gasJSON},
	} {

		if entry.Type == "gas" {
			if tele.Gas.Time == prevGasTimestamp {
				continue
			}
		}

		_, err := db.Exec(`
			INSERT INTO iot_data (time, device_id, type, data)
			VALUES ($1, $2, $3, $4)`,
			tele.Timestamp, deviceID, entry.Type, entry.Data)

		log.Printf("Record written!")
		if err != nil {
			return fmt.Errorf("insert %s: %w", entry.Type, err)
		}
	}

	return nil
}
