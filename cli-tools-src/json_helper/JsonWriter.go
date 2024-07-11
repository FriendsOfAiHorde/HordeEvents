package json_helper

import (
	"encoding/json"
	"os"
)

func WriteJson(filename string, value any) error {
	jsonRaw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, jsonRaw, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func WriteJsonMinified(filename string, value any) error {
	jsonRaw, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, jsonRaw, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
