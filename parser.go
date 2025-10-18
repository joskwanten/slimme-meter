package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

type Value struct {
	Value string `json:"value"`
	Unit  string `json:"unit,omitempty"`
}

type Telegram struct {
	Timestamp time.Time          `json:"timestamp"`
	Values    map[string][]Value `json:"values"` // key = OBIS-code
}

var obisMap = map[string]string{
	// Elektriciteit - geleverde energie (kWh)
	"1-0:1.8.1": "Delivered.Tariff1",
	"1-0:1.8.2": "Delivered.Tariff2",

	// Elektriciteit - teruggeleverde energie (kWh)
	"1-0:2.8.1": "Returned.Tariff1",
	"1-0:2.8.2": "Returned.Tariff2",

	// Vermogen (kW)
	"1-0:1.7.0": "ActivePower",
	"1-0:2.7.0": "ReactivePower",

	// Stroom per fase (A)
	"1-0:31.7.0": "Current.L1",
	"1-0:51.7.0": "Current.L2",
	"1-0:71.7.0": "Current.L3",

	// Spanning per fase (V)
	"1-0:32.7.0": "Voltage.L1",
	"1-0:52.7.0": "Voltage.L2",
	"1-0:72.7.0": "Voltage.L3",

	// Gas (m3)
	"0-1:24.2.1": "Gas.Volume",

	// Optioneel: meter ID / serienummer
	"0-0:96.1.1": "Electricity.MeterID",
	"0-1:96.1.0": "Gas.MeterID",

	// Huidige tijd / timestamp van telegram
	"0-0:1.0.0": "Telegram.Timestamp",
}

func parseRawValue(raw string) Value {
	re := regexp.MustCompile(`^(?P<value>[^*]+)(?:\*(?P<unit>.+))?$`)
	match := re.FindStringSubmatch(raw)
	names := re.SubexpNames()

	var value, unit string
	for i, name := range names {
		if i != 0 && name != "" {
			switch name {
			case "value":
				value = match[i]
			case "unit":
				unit = match[i]
			}
		}
	}

	return Value{
		Value: value,
		Unit:  unit,
	}
}

func parseRawValues(raw string) []Value {
	var values []Value

	re := regexp.MustCompile(`\((?P<value>[^\)]+)\)`)
	matches := re.FindAllStringSubmatch(raw, -1)

	for _, m := range matches {
		value := parseRawValue(m[1])
		fmt.Printf("%s %s\n", value.Value, value.Unit)
		values = append(values, value)
	}

	return values
}

func main() {
	// Open je testbestand
	file, err := os.Open("testfile.txt")
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	t := Telegram{
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
			fmt.Printf("%s: %s\n", fieldName, valuesPart)
			values := parseRawValues(valuesPart)
			t.Values[obisMap[obis]] = values
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	bytes, err := json.Marshal(t)
	log.Printf("JSON: %s", bytes)
}
