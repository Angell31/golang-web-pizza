package helper

import (
	"encoding/json"
	"log"
	"os"
	"pizza/data"
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

func GetPizzaById(id int) (*data.Pizza, int) {
	for i, o := range data.Orders {
		if o.ID == id {
			return &data.Menu[i], i
		}
	}
	return nil, 0
}
