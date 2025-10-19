package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tarm/serial"
	// zie uitleg hieronder
)

func main() {
	portName := "/dev/ttyUSB0"
	baud := 115200

	c := &serial.Config{Name: portName, Baud: baud}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatalf("failed to open serial port: %v", err)
	}
	defer s.Close()

	scanner := bufio.NewScanner(s)

	DSMRScanner(scanner, printTelegram)

	if err := scanner.Err(); err != nil {
		log.Printf("error reading from serial: %v", err)
	}
}

func printTelegram(telegram Telegram) {
	bytes, _ := json.Marshal(telegram)
	fmt.Printf("JSON: %s", bytes)
}
