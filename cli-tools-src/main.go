package main

import (
	"flag"
	"fmt"
	"github.com/araddon/dateparse"
	"github.com/google/uuid"
	"github.com/xeipuuv/gojsonschema"
	"main/generator"
	"main/helper"
	"main/horde"
	"main/parser"
	"os"
	"slices"
	"strings"
	"time"
)

func main() {
	commands := map[string]map[string]map[string]string{
		"add": {
			"title": {
				"description": "The name of the event",
				"required":    "y",
			},
			"description": {
				"description": "The longer text of the notification, describing the event",
			},
			"link": {
				"description": "A link that the notification will point to",
			},
			"valid-since": {
				"description": "The date and time where the event starts, notifications won't be shown before this date",
				"required":    "y",
			},
			"valid-until": {
				"description": "The date and time where the event starts, notifications won't be shown before this date",
				"required":    "y",
			},
		},
		"remove": {
			"id": {
				"description": "The ID if the event you want to remove",
				"required":    "y",
			},
		},
		"cleanup":  {},
		"format":   {},
		"validate": {},
		"generate": {},
	}

	var clients []string
	err := helper.MapJson(getClientsFileName(), &clients)
	commands["add"]["only"] = make(map[string]string)
	if err != nil {
		commands["add"]["only"]["description"] = "Specify the names of the projects this applies to (separated by comma)"
	} else {
		commands["add"]["only"]["description"] = "Specify the names of the projects this applies to (separated by comma), you can use one of '" + strings.Join(clients, "', '") + "', but any other string is valid as well"
	}

	channelNames, err := parser.NewSchemaParser(getSchemaFileName()).GetAllowedChannels()
	commands["add"]["channels"] = make(map[string]string)
	if err != nil {
		commands["add"]["channels"]["description"] = "Specify the channel names, check schema.json file for valid values (separated by comma)"
	} else {
		commands["add"]["channels"]["description"] = "Specify the channel names, valid values: " + strings.Join(channelNames, ", ")
	}

	values := make(map[string]map[string]*string)
	flagSets := make(map[string]*flag.FlagSet)

	for topCommand := range commands {
		flagSets[topCommand] = flag.NewFlagSet(topCommand, flag.ExitOnError)
		values[topCommand] = make(map[string]*string)
		for flagName := range commands[topCommand] {
			config := commands[topCommand][flagName]
			values[topCommand][flagName] = flagSets[topCommand].String(flagName, "", config["description"])
		}
	}

	if len(os.Args) < 2 {
		printHelp("", &commands)
		os.Exit(1)
	}

	topCommands := make([]string, 0, len(commands))
	for key := range commands {
		topCommands = append(topCommands, key)
	}

	if !slices.Contains(topCommands, os.Args[1]) {
		printHelp(os.Args[1], &commands)
		os.Exit(1)
	}

	command := os.Args[1]
	flagSets[command].Parse(os.Args[2:])
	config := commands[command]

	for flagName := range values[command] {
		config[flagName]["value"] = *values[command][flagName]
	}

	if !validate(&config) {
		printHelp(command, &commands)
		os.Exit(1)
	}

	switch command {
	case "add":
		os.Exit(handleAdd(config))
	case "remove":
		os.Exit(handleRemove(config))
	case "cleanup":
		os.Exit(handleCleanup())
	case "format":
		os.Exit(handleFormat())
	case "validate":
		os.Exit(handleValidate())
	case "generate":
		os.Exit(handleGenerate())
	default:
		fmt.Println("Unhandled command:", command)
		os.Exit(1)
	}
}

func handleGenerate() int {
	clients := make([]string, 0)
	err := helper.MapJson(getClientsFileName(), &clients)
	if err != nil {
		fmt.Println(err)
		return 1
	}

	resultGenerator := generator.NewResultGenerator(clients, getJson())
	err = resultGenerator.Generate()
	if err != nil {
		fmt.Println(err)
		return 1
	}

	fmt.Println("The resulting files have been generated")
	return 0
}

func handleValidate() int {
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + getSchemaFileName())
	documentLoader := gojsonschema.NewReferenceLoader("file://" + getJsonFileName())

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)

	if err != nil {
		fmt.Println("There was an error while validating the JSON schema:")
		fmt.Println(err)
		return 1
	}

	if !result.Valid() {
		fmt.Println("The JSON schema is not valid:")
		for _, desc := range result.Errors() {
			fmt.Println("-", desc)
		}
		return 1
	}

	fmt.Println("The source.json file is valid")
	return 0
}

func handleFormat() int {
	writeJson(getJson())
	fmt.Println("The source.json file has been successfully formatted")
	return 0
}

func printHelp(command string, config *map[string]map[string]map[string]string) {
	topCommands := make([]string, 0, len(*config))
	for key := range *config {
		topCommands = append(topCommands, key)
	}

	if command == "" {
		fmt.Println("You must use one of the available subcommands:", strings.Join(topCommands, ", "))
		return
	}

	if !slices.Contains(topCommands, command) {
		fmt.Println(command, "is not a valid subcommand, please use one of the following:", strings.Join(topCommands, ", "))
		return
	}

	commandConfig := (*config)[command]
	fmt.Println("Usage:")
	fmt.Print(os.Args[0] + " " + command)

	flagNames := make([]string, 0, len(commandConfig))
	for key := range commandConfig {
		flagNames = append(flagNames, key)
	}

	slices.SortFunc(flagNames, func(a, b string) int {
		configA := commandConfig[a]
		configB := commandConfig[b]

		if configA["required"] == "y" && configB["required"] == "y" {
			return 0
		}

		if configA["required"] == "y" {
			return -1
		}

		return 1
	})

	for _, flagName := range flagNames {
		flagConfig := commandConfig[flagName]
		fmt.Print(" ")
		if flagConfig["required"] != "y" {
			fmt.Print("[")
		}
		fmt.Print(" ")

		fmt.Print("--", flagName)
		fmt.Print(" ")

		fmt.Print("<value>")

		if flagConfig["required"] != "y" {
			fmt.Print(" ]")
		}
	}

	fmt.Println()
	fmt.Println()
	fmt.Println("Options:")

	for _, flagName := range flagNames {
		flagConfig := commandConfig[flagName]

		fmt.Print("  --"+flagName, "\t\t")
		if flagConfig["required"] == "y" {
			fmt.Print("(Required)")
		} else {
			fmt.Print("(Optional)")
		}
		fmt.Print("\t")
		fmt.Print(flagConfig["description"])

		fmt.Println()
	}
}

func validate(config *map[string]map[string]string) bool {
	for key := range *config {
		fieldConfig := (*config)[key]
		if fieldConfig["required"] == "y" && fieldConfig["value"] == "" {
			return false
		}
	}

	return true
}

func handleAdd(config map[string]map[string]string) int {
	name := config["title"]["value"]
	description := config["description"]["value"]
	link := config["link"]["value"]
	only := config["only"]["value"]
	channels := config["channels"]["value"]
	validSince, err := dateparse.ParseAny(config["valid-since"]["value"])
	if err != nil {
		fmt.Println("Failed parsing valid-since date:")
		fmt.Println(err)
		return 1
	}
	validUntil, err := dateparse.ParseAny(config["valid-until"]["value"])
	if err != nil {
		fmt.Println("Failed parsing valid-until date:")
		fmt.Println(err)
		return 1
	}

	utc, err := time.LoadLocation("UTC")
	if err != nil {
		fmt.Println("Failed loading UTC location")
		return 1
	}

	event := horde.Event{
		Id:          uuid.New(),
		Title:       name,
		ValidSince:  validSince.In(utc),
		ValidUntil:  validUntil.In(utc),
		Description: nil,
		LimitedTo:   nil,
		Link:        nil,
		Channels:    nil,
	}
	if description != "" {
		event.Description = &description
	}
	if only != "" {
		event.LimitedTo = strings.Split(only, ",")
	}
	if link != "" {
		event.Link = &link
	}
	if channels != "" {
		event.Channels = strings.Split(channels, ",")

		availableChannels, err := parser.NewSchemaParser(getSchemaFileName()).GetAllowedChannels()
		if err != nil {
			fmt.Println("Failed getting list of allowed channels:")
			fmt.Println(err)
			return 1
		}
		for _, channel := range event.Channels {
			if !slices.Contains(availableChannels, channel) {
				fmt.Println("The value '"+channel+"' is not a valid channel, valid values include:", strings.Join(availableChannels, ", "))
				return 1
			}
		}
	}

	if addJson(event) {
		fmt.Println("The event has been successfully added")
		return 0
	}

	fmt.Println("Failed to create the event")
	return 0
}

func handleRemove(config map[string]map[string]string) int {
	jsonArray := getJson()
	if jsonArray == nil {
		return 1
	}

	id := config["id"]["value"]

	found := false
	for index, item := range jsonArray {
		if item.Id.String() == id {
			found = true
			jsonArray = append(jsonArray[:index], jsonArray[index+1:]...)
			break
		}
	}

	writeJson(jsonArray)

	if found {
		fmt.Println("The event has been successfully removed")
		return 0
	}
	fmt.Println("Event with id", id, "was not found")
	return 1
}

func handleCleanup() int {
	jsonArray := getJson()

	now := time.Now()

	removed := 0
	for currentPassRemoved := true; currentPassRemoved; {
		currentPassRemoved = false
		for index, item := range jsonArray {
			if now.After(item.ValidUntil) {
				jsonArray = append(jsonArray[:index], jsonArray[index+1:]...)
				removed += 1
				currentPassRemoved = true
				break
			}
		}
	}

	if removed > 0 {
		fmt.Println("Successfully removed", removed, "expired events")
		writeJson(jsonArray)
	} else {
		fmt.Println("No expired events found")
	}

	return 0
}

func addJson(data horde.Event) bool {
	result := getJson()
	if result == nil {
		return false
	}

	result = append(result, data)
	writeJson(result)

	return true
}

func getJson() []horde.Event {
	result := make([]horde.Event, 0)
	err := helper.MapJson(getJsonFileName(), &result)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return result
}

func writeJson(jsonArray []horde.Event) {
	err := helper.WriteJson(getJsonFileName(), jsonArray)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func getJsonFileName() string {
	return "./source.json"
}

func getSchemaFileName() string {
	return "./schema.json"
}

func getClientsFileName() string {
	return "./clients.json"
}
