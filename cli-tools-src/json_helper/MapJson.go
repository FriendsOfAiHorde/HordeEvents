package json_helper

import (
	"encoding/json"
	"io"
	"os"
)

func MapJson[TOutput any](filename string, output *TOutput) error {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer jsonFile.Close()

	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, output)
	if err != nil {
		return err
	}

	return nil
}
