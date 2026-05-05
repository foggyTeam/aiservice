package main

import (
	"encoding/json"
	"os"
)

func ToJSON(req any) ([]byte, error) {
	return json.MarshalIndent(req, "", " ")
}

func SaveToFile(req any, filename string) error {
	data, err := ToJSON(req)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func main() {
	runRequestGenerator()
}
