package main

import (
	"bufio"
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
)

// --- Structs voor type-safe JSONB ---
type Electricity struct {
	Tariff1          float64 `json:"tariff1,omitempty"`
	Tariff2          float64 `json:"tariff2,omitempty"`
	ExportTariff1    float64 `json:"export_tariff1,omitempty"`
	PowerDeliveredKW float64 `json:"power_delivered_kW,omitempty"`
}

type Measurement struct {
	Timestamp   time.Time   `json:"timestamp"`
	Electricity Electricity `json:"electricity,omitempty"`
	Gas         float64     `json:"gas,omitempty"`
}

// --- Config ---
type Config struct {
	SerialPort string
	Baud       int
	PGDSN      string
	DeviceID   string
	Debug      bool
}

func main() {
	portName := "/dev/ttyUSB0" // pas aan naar jouw device
	baud := 115200

	c := &serial.Config{Name: portName, Baud: baud}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatalf("failed to open serial port: %v", err)
	}
	defer s.Close()

	scanner := bufio.NewScanner(s)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Printf("error reading from serial: %v", err)
	}
}
