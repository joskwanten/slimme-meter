package main

import (
	"fmt"
	"github.com/tarm/serial"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func crc16(buf []byte) uint {
	var crc uint = 0
	for pos := 0; pos < len(buf); pos++ {
		crc ^= uint(buf[pos]) // XOR byte into least sig. byte of crc

		for i := 8; i != 0; i-- { // Loop over each bit
			if (crc & 0x0001) != 0 { // If the LSB is set
				crc >>= 1 // Shift right and XOR 0xA001
				crc ^= 0xA001
			} else {
				// Else LSB is not set
				crc >>= 1 // Just shift right
			}
		}
	}
	return crc
}

func ReadMessage(msgchan chan []byte) {
	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 115200}
	s, err := serial.OpenPort(c)

	if err != nil {
		log.Fatal(err)
	}

	for {
		startFound := false
		endFound := false
		message := []byte{}

		for !endFound {
			buf := make([]byte, 2000)
			n, err := s.Read(buf)
			if err != nil {
				log.Fatal(err)
			}

			for i := 0; i < n; i++ {
				if buf[i] == '!' {
					endFound = true
					message = append(message, buf[0:i+1]...)
					crcStr := strings.Replace(string(buf[i+1:n]), "\r\n", "", -1)
					crc, err := strconv.ParseInt("0x"+crcStr, 0, 32)
					crcComp := int64(crc16(message))

					if err == nil && crc == crcComp {
						msgchan <- message
					} else {
						log.Printf("Failed to parse message")
					}
					break
				}
			}

			if !startFound {
				for i := 0; i < n; i++ {
					if buf[i] == '/' {
						startFound = true
						message = append(buf[i : n-i])
					}
				}
			} else {
				message = append(message, buf[0:n]...)
			}
		}
	}

}

func main() {
	r := regexp.MustCompile(`(?P<device>\d+-\d+):(?P<key>\d+.\d+.\d+)(?P<values>(\(.*?\))+)`)
	valueRegex := regexp.MustCompile(`\((\d*.\d*)\*kWh\)`)
	timeRegex := regexp.MustCompile(`\((\d*)W\)`)
	gasRegex := regexp.MustCompile(`\((\d*)W\)\((\d+.\d+)\*m3\)`)
	messages := make(chan []byte)
	go ReadMessage(messages)
	for {
		var tarief1Afgenomen, tarief2Afgenomen, tarief1Teruggeleverd, tarief2Teruggeleverd, gasVerbruik float64
		var timestamp, gasTimestamp time.Time
		message := <-messages
		parts := strings.Split(string(message), "\r\n")
		for i := 0; i < len(parts); i++ {
			res := r.FindStringSubmatch(parts[i])
			if len(res) >= 3 {
				switch res[2] {
				case "1.0.0":
					valResult := timeRegex.FindStringSubmatch(res[3])
					if len(valResult) == 2 {
						timestamp, _= time.Parse("20060102150405", "20" + valResult[1])
					}
				case "1.8.1":
					valResult := valueRegex.FindStringSubmatch(res[3])
					if len(valResult) == 2 {
						tarief1Afgenomen, _ = strconv.ParseFloat(valResult[1], 64)
					}
				case "1.8.2":
					valResult := valueRegex.FindStringSubmatch(res[3])
					if len(valResult) == 2 {
						tarief2Afgenomen, _ = strconv.ParseFloat(valResult[1], 64)
					}
				case "2.8.1":
					valResult := valueRegex.FindStringSubmatch(res[3])
					if len(valResult) == 2 {
						tarief1Teruggeleverd, _ = strconv.ParseFloat(valResult[1], 64)
					}
				case "2.8.2":
					valResult := valueRegex.FindStringSubmatch(res[3])
					if len(valResult) == 2 {
						tarief2Teruggeleverd, _ = strconv.ParseFloat(valResult[1], 64)
					}
				case "24.2.1":
					valResult := gasRegex.FindStringSubmatch(res[3])
					if len(valResult) == 3 {
						gasTimestamp, _= time.Parse("20060102150405", "20" + valResult[1])
						gasVerbruik, _ = strconv.ParseFloat(valResult[2], 64)
					}

					
				}

				if res[2] == "1.0.0" {
					fmt.Printf("Tijd: %s\n", timestamp.Format("20060102150405"))
				} else if res[2] == "1.8.1" {
					fmt.Printf("Totaalverbruik Tarief 1 (nacht): %f\n", tarief1Afgenomen)
				} else if res[2] == "1.8.2" {
					fmt.Printf("Totaalverbruik Tarief 2 (dag): %f\n", tarief2Afgenomen)
				} else if res[2] == "2.8.1" {
					fmt.Printf("Totaal geleverd Tarief 1 (nacht): %f\n", tarief1Teruggeleverd)
				} else if res[2] == "2.8.2" {
					fmt.Printf("Totaal geleverd Tarief 2 (dag): %f\n", tarief2Teruggeleverd)
				} else if res[2] == "24.2.1" {
					fmt.Printf("Gas timestamp %s\n", gasTimestamp.Format("20060102150405"))
					fmt.Printf("Gasverbruik: %f\n", gasVerbruik)
				}
			}
		}
	}
	//log.Printf("%d", n)
}
