package helper

import (
	"encoding/json"
	"log"
	"os"
)

func ReadJSONFile(filename string, v interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open menu file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	if err := decoder.Decode(&v); err != nil {
		log.Fatalf("Failed to parse menu file: %v", err)
	}
	return nil
}
