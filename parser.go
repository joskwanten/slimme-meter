package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Value struct {
	Value string `json:"value"`
	Unit  string `json:"unit,omitempty"`
}

type RawTelegram struct {
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

type Electricity struct {
	Delivered struct {
		Tariff1 float64 `json:"Tariff1"` // kWh
		Tariff2 float64 `json:"Tariff2"` // kWh
	} `json:"Delivered"`
	Returned struct {
		Tariff1 float64 `json:"Tariff1"` // kWh
		Tariff2 float64 `json:"Tariff2"` // kWh
	} `json:"Returned"`
	ActivePower   float64 `json:"ActivePower"`   // kW
	ReactivePower float64 `json:"ReactivePower"` // kW
	Current       struct {
		L1 float64 `json:"L1"` // A
		L2 float64 `json:"L2"` // A
		L3 float64 `json:"L3"` // A
	} `json:"Current"`
}

type Gas struct {
	MeterID string    `json:"MeterID"`
	Volume  float64   `json:"Volume"` // m3
	Time    time.Time `json:"Time"`   // meetmoment
}

type Telegram struct {
	Timestamp   time.Time   `json:"timestamp"` // telegram ontvangsttijd
	Electricity Electricity `json:"electricity"`
	Gas         Gas         `json:"gas"`
}

// parseFloat zet een string om naar float64, bij fouten return 0
func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Printf("parseFloat error: %v (input: %s)", err, s)
		return 0
	}
	return f
}

// parseTime zet een DSMR-timestamp string om naar time.Time
// DSMR timestamp formaat: YYMMDDhhmmssS, bijv "251018095310S"
func parseTime(s string) time.Time {
	t, err := time.Parse("060102150405S", s)
	if err != nil {
		log.Printf("parseTime error: %v (input: %s)", err, s)
		return time.Time{}
	}
	return t
}

func MapRawTelegramDynamic(raw RawTelegram) (Telegram, error) {
	var t Telegram
	t.Timestamp = raw.Timestamp

	// Elektriciteit en Gas structen tijdelijk
	var elec Electricity
	var gas Gas

	for obis, values := range raw.Values {
		if len(values) == 0 {
			continue
		}

		// Kies eerste Value (meestal is er 1, behalve bij gas: [timestamp, volume])
		v := values[0].Value

		switch obis {
		// ---------------------------
		// Elektriciteit
		// ---------------------------
		case "Delivered.Tariff1":
			elec.Delivered.Tariff1 = parseFloat(v)
		case "Delivered.Tariff2":
			elec.Delivered.Tariff2 = parseFloat(v)
		case "Returned.Tariff1":
			elec.Returned.Tariff1 = parseFloat(v)
		case "Returned.Tariff2":
			elec.Returned.Tariff2 = parseFloat(v)
		case "ActivePower":
			elec.ActivePower = parseFloat(v)
		case "ReactivePower":
			elec.ReactivePower = parseFloat(v)
		case "Current.L1":
			elec.Current.L1 = parseFloat(v)
		case "Current.L2":
			elec.Current.L2 = parseFloat(v)
		case "Current.L3":
			elec.Current.L3 = parseFloat(v)
		case "Voltage.L1":
			// elec.Voltage.L1 = parseFloat(v) // voeg Voltage substruct toe als nodig
		case "Voltage.L2":
		case "Voltage.L3":
			// ...
		case "Electricity.MeterID":
			elecMeterID := v // alleen string
			// je kunt evt. een field toevoegen in Electricity struct
			_ = elecMeterID

		// ---------------------------
		// Gas
		// ---------------------------
		case "Gas.MeterID":
			gas.MeterID = v
		case "Gas.Volume":
			if len(values) > 1 {
				gas.Time = parseTime(values[0].Value)
				gas.Volume = parseFloat(values[1].Value)
			} else {
				gas.Volume = parseFloat(v)
			}

		// ---------------------------
		// Telegram timestamp
		// ---------------------------
		case "Telegram.Timestamp":
			t.Timestamp = parseTime(v)
		}
	}

	t.Electricity = elec
	t.Gas = gas
	return t, nil
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

	t := RawTelegram{
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
			t.Values[fieldName] = values
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	bytes, err := json.Marshal(t)
	log.Printf("JSON: %s", bytes)

	t2, _ := MapRawTelegramDynamic(t)
	bytes2, err := json.Marshal(t2)
	log.Printf("JSON: %s", bytes2)

}
