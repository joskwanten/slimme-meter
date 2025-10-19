package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

func printTestTelegram(telegram Telegram) {
	bytes, _ := json.Marshal(telegram)
	fmt.Printf("JSON: %s", bytes)
}

func TestParser(t *testing.T) {
	// Open je testbestand
	file, err := os.Open("testfile.txt")
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	rawTelegram := RawTelegram{
		Timestamp: time.Now().UTC(),
		Values:    make(map[string][]Value),
	}

	// Scanner voor regel-voor-regel lezen
	scanner := bufio.NewScanner(file)
	DSMRScanner(scanner, printTestTelegram)

	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	telegram, _ := MapRawTelegramDynamic(rawTelegram)
	bytes, _ := json.Marshal(telegram)
	log.Printf("JSON: %s", bytes)

}
