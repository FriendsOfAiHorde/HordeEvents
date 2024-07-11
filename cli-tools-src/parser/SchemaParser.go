package parser

import (
	"main/helper"
	"os"
)

type schema struct {
	Items struct {
		Properties struct {
			Channels struct {
				Items struct {
					Enum []string
				}
			}
		}
	}
}

type SchemaParser struct {
	filepath string
}

func NewSchemaParser(filepath string) *SchemaParser {
	return &SchemaParser{
		filepath: filepath,
	}
}

func (parser SchemaParser) openForReading() (*os.File, error) {
	return os.Open(parser.filepath)
}

func (parser SchemaParser) GetAllowedChannels() ([]string, error) {
	var result schema
	err := helper.MapJson(parser.filepath, &result)
	if err != nil {
		return nil, err
	}

	return result.Items.Properties.Channels.Items.Enum, nil
}
