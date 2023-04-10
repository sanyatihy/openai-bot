package processor

import (
	"encoding/json"
	"errors"
	"os"
)

func (p *processor) loadLastUpdateIDFromFile(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	defer file.Close()

	data := make(map[string]int)
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		return 0, err
	}

	return data["lastUpdateID"], nil
}

func (p *processor) saveLastUpdateIDToFile(filename string, lastUpdateID int) error {
	data := map[string]int{"lastUpdateID": lastUpdateID}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}
