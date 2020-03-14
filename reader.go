package main

import (
	"strconv"
	"fmt"
	"github.com/tarm/serial"
	"log"
)

func crc16(buf []byte) uint {
	var crc uint = 0
	for pos:=0; pos<len(buf); pos++ {
	  crc ^= uint(buf[pos])  // XOR byte into least sig. byte of crc
	  
	  for i :=8; i != 0; i-- {       // Loop over each bit
		if (crc & 0x0001) != 0 {    // If the LSB is set
		  crc >>= 1;                  // Shift right and XOR 0xA001
		  crc ^= 0xA001;
		} else {
			                 // Else LSB is not set
		  crc >>= 1;                  // Just shift right
		}
	  }
	}
	return crc;
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
					message = append(message, buf[0:i + 1]...)
					crc, err := strconv.ParseInt("0x"+string(buf[i+1:n-2]), 0, 32)
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
						message = append(buf[i:n-i])
					}
				}
			} else {
				message = append(message, buf[0:n]...)
			}
		}
	}

}

func main() {
	messages := make(chan []byte)
	go ReadMessage(messages)
	for {
		message := <-messages

		fmt.Printf("%s", message)
	}
	//log.Printf("%d", n)
}
