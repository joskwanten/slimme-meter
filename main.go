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
	idx := strings.LastIndex(telegram, "!")
	if idx == -1 || idx+5 > len(telegram) {
		return false
	}

	data := telegram[:idx]
	expectedHex := telegram[idx+1 : idx+5]

	crc := calcCRC16([]byte(data))
	actualHex := fmt.Sprintf("%04X", crc)

	return strings.EqualFold(expectedHex, actualHex)
}

// calcCRC16 berekent de CRC16/X25 (DSMR gebruikt CRC16/X25)
func calcCRC16(data []byte) uint16 {
	const poly = uint16(0x1021)
	const init = uint16(0xFFFF)
	const xorOut = uint16(0xFFFF)

	crc := init
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if (crc & 0x8000) != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc <<= 1
			}
		}
	}
	return crc ^ xorOut
}
