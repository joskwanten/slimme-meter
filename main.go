package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"strings"

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
	var buffer bytes.Buffer
	inTelegram := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "/") {
			// Nieuw telegram begint
			buffer.Reset()
			inTelegram = true
		}

		if inTelegram {
			buffer.WriteString(line + "\n")

			if strings.HasPrefix(line, "!") {
				inTelegram = false
				raw := buffer.String()
				if validateCRC(raw) {
					fmt.Println("✅ Valid telegram:\n", raw)
				} else {
					fmt.Println("❌ Invalid checksum:\n", raw)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("error reading from serial: %v", err)
	}
}

// validateCRC controleert de CRC16-checksum volgens DSMR P1-protocol
func validateCRC(telegram string) bool {
	// Vind ! en vier hex digits erna
	idx := strings.LastIndex(telegram, "!")
	if idx == -1 || idx+5 > len(telegram) {
		return false
	}

	// Data vóór het '!' zonder trailing CR/LF
	data := strings.TrimRight(telegram[:idx], "\r\n")

	// verwacht hex (4 chars) direct na '!'
	expectedHex := telegram[idx+1 : idx+5]

	crc := calcCRC16X25([]byte(data))
	actualHex := fmt.Sprintf("%04X", crc)

	// debug (haal weg als het werkt)
	// fmt.Printf("DATA BYTES: % X\n", []byte(data))
	// fmt.Printf("Expected: %s, Actual: %s\n", expectedHex, actualHex)

	return strings.EqualFold(expectedHex, actualHex)
}

// calcCRC16X25 berekent CRC-16/X-25 (LSB-first, reflected poly 0x1021 as 0x8408)
func calcCRC16X25(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if (crc & 0x0001) != 0 {
				crc = (crc >> 1) ^ 0x8408
			} else {
				crc >>= 1
			}
		}
	}
	// final XOR
	return ^crc
}
