package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tarm/serial"
	// zie uitleg hieronder
)

func main() {
	portName := "/dev/ttyUSB0"
	baud := 115200

	// PostgreSQL DSN (pas aan)
	dsn := "host=localhost port=30042 user=postgres password=postgres dbname=home sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	defer db.Close()

	Migrate(db)

	c := &serial.Config{Name: portName, Baud: baud}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatalf("failed to open serial port: %v", err)
	}
	defer s.Close()

	scanner := bufio.NewScanner(s)

	var prevTelegram Telegram

	storeData := func(telegram Telegram) {
		bytes, _ := json.Marshal(telegram)
		fmt.Printf("Storing : %s", bytes)
		StoreTelegram(db, "home", telegram, prevTelegram.Gas.Time)
		prevTelegram = telegram
	}

	DSMRScanner(scanner, storeData)

	if err := scanner.Err(); err != nil {
		log.Printf("error reading from serial: %v", err)
	}
}
