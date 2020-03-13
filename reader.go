package main

import (
      "github.com/tarm/serial"
      "log"
)

func main() {
    c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 115200}
    s, err := serial.OpenPort(c)
	
	if err != nil {
        log.Fatal(err)
    }

	for {
    	buf := make([]byte, 2000)
      	n, err := s.Read(buf)
      	if err != nil {
              log.Fatal(err)
      	}
		
		log.Printf("%s", buf[:n])
		//log.Printf("%d", n)
		  
	}
}