package generator

import (
	"fmt"
	"main/helper"
	"main/horde"
	"slices"
	"time"
)

type ResultGenerator struct {
	staticClients []string
	events        []horde.Event
}

func NewResultGenerator(
	staticClients []string,
	events []horde.Event,
) *ResultGenerator {
	return &ResultGenerator{
		staticClients: staticClients,
		events:        events,
	}
}

func (generator ResultGenerator) Generate() error {
	if generator.staticClients == nil || generator.events == nil {
		return fmt.Errorf("please use the NewResultGenerator function to create an instance of generator")
	}

	utc, err := time.LoadLocation("UTC")
	if err != nil {
		fmt.Println("Failed loading UTC location")
		return err
	}

	const COMMON = "common"

	now := time.Now()
	ids := make(map[string]bool, len(generator.events))
	results := make(map[string][]horde.Event)

	for _, item := range generator.events {
		item.ValidSince = item.ValidSince.In(utc)
		item.ValidUntil = item.ValidUntil.In(utc)

		if _, exists := ids[item.Id.String()]; exists {
			return fmt.Errorf("the ID '%s' already exists", item.Id)
		}
		ids[item.Id.String()] = true

		if now.Before(item.ValidSince) || now.After(item.ValidUntil) {
			continue
		}
		if item.LimitedTo == nil {
			item.LimitedTo = make([]string, 1)
			item.LimitedTo[0] = COMMON
		}
		for _, limitedTo := range item.LimitedTo {
			if _, exists := results[limitedTo]; !exists {
				results[limitedTo] = make([]horde.Event, 0)
			}
			results[limitedTo] = append(results[limitedTo], item)
		}
	}

	for _, client := range generator.staticClients {
		if _, exists := results[client]; !exists {
			results[client] = make([]horde.Event, 0)
		}
	}
	if _, exists := results[COMMON]; !exists {
		results[COMMON] = make([]horde.Event, 0)
	}

	for key, items := range results {
		if key == COMMON {
			continue
		}

		results[key] = append(items, results[COMMON]...)
	}

	for key, _ := range results {
		slices.SortStableFunc(results[key], func(a, b horde.Event) int {
			if a.ValidSince.After(b.ValidSince) {
				return -1
			}
			if a.ValidSince.Before(b.ValidSince) {
				return 1
			}

			return 0
		})
	}

	for limitedTo, result := range results {
		filename := fmt.Sprintf("results.%s.json", limitedTo)
		filenameMin := fmt.Sprintf("results.%s.min.json", limitedTo)

		err1 := helper.WriteJson(filename, result)
		err2 := helper.WriteJsonMinified(filenameMin, result)

		if err1 != nil {
			return err1
		}
		if err2 != nil {
			return err2
		}
	}

	return nil
}
