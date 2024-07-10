package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/araddon/dateparse"
	"github.com/google/uuid"
	"io"
	"os"
	"slices"
	"strings"
	"time"
)

type HordeEvent struct {
	Id          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	ValidSince  time.Time `json:"validSince"`
	ValidUntil  time.Time `json:"validUntil"`
	Description *string   `json:"description,omitempty"`
	LimitedTo   []string  `json:"limitedTo,omitempty"`
	Link        *string   `json:"link,omitempty"`
	Channels    []string  `json:"channels,omitempty"`
}

func main() {
	commands := make(map[string]map[string]map[string]string)
	commands["add"] = make(map[string]map[string]string)
	commands["remove"] = make(map[string]map[string]string)
	commands["cleanup"] = make(map[string]map[string]string)

	commands["add"]["title"] = make(map[string]string)
	commands["add"]["title"]["description"] = "The name of the event"
	commands["add"]["title"]["required"] = "y"

	commands["add"]["description"] = make(map[string]string)
	commands["add"]["description"]["description"] = "The longer text of the notification, describing the event"

	commands["add"]["link"] = make(map[string]string)
	commands["add"]["link"]["description"] = "A link that the notification will point to"

	commands["add"]["valid-since"] = make(map[string]string)
	commands["add"]["valid-since"]["description"] = "The date and time where the event starts, notifications won't be shown before this date"
	commands["add"]["valid-since"]["required"] = "y"

	commands["add"]["valid-until"] = make(map[string]string)
	commands["add"]["valid-until"]["description"] = "The date and time where the event starts, notifications won't be shown after this date"
	commands["add"]["valid-until"]["required"] = "y"

	commands["add"]["only"] = make(map[string]string)
	commands["add"]["only"]["description"] = "Specify the names of the projects this applies to (separated by comma)"

	commands["add"]["channels"] = make(map[string]string)
	commands["add"]["channels"]["description"] = "Specify the channel names, check schema.json file for valid values (separated by comma)"

	commands["remove"]["id"] = make(map[string]string)
	commands["remove"]["id"]["description"] = "The ID if the event you want to remove"
	commands["remove"]["id"]["required"] = "y"

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
	default:
		fmt.Println("Unhandled command:", command)
		os.Exit(1)
	}
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

	event := HordeEvent{
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

func addJson(data HordeEvent) bool {
	result := getJson()
	if result == nil {
		return false
	}

	result = append(result, data)
	writeJson(result)

	return true
}

func getJson() []HordeEvent {
	jsonFile, err := os.Open(getJsonFileName())
	if err != nil {
		fmt.Println(err)
		return nil
	}

	defer jsonFile.Close()

	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	result := make([]HordeEvent, 0)
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return result
}

func writeJson(jsonArray []HordeEvent) {
	jsonRaw, err := json.MarshalIndent(jsonArray, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}

	os.WriteFile(getJsonFileName(), jsonRaw, os.ModePerm)
}

func getJsonFileName() string {
	return "source.json"
}
