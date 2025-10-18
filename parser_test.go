package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

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
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // lege regels overslaan
		}

		// --- Hier kun je de parsing doen ---
		// Bijvoorbeeld OBIS-code + value/unit als string
		i := strings.Index(line, "(")
		if i < 0 {
			continue
		}

		obis := line[:i]
		valuesPart := line[i:] // alles inclusief haakjes, of gebruik line[i+1:len(line)-1] voor alleen binnen haakjes
		if fieldName, ok := obisMap[obis]; ok {
			values := parseRawValues(valuesPart)
			rawTelegram.Values[fieldName] = values
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	telegram, _ := MapRawTelegramDynamic(rawTelegram)
	bytes, _ := json.Marshal(telegram)
	log.Printf("JSON: %s", bytes)

}
