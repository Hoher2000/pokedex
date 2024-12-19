package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Hoher2000/pokedex/pokecache"
)

type cliCommand struct {
	name        string
	description string
	operation   func(c *config) error
}

type pokeLocation struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type pokeRespLocation struct {
	Count    int            `json:"count"`
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []pokeLocation `json:"results"`
}

type config struct {
	previousUrl string
	nextUrl     string
}

func pokeRequest(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if res.StatusCode > 299 {
		return nil, fmt.Errorf("response failed with status code: %d and\nbody: %s", res.StatusCode, body)
	}
	if err != nil {
		return nil, err
	}
	return body, nil
}

var cache = pokecache.NewCache(time.Second * 5)

func main() {
	var commands map[string]cliCommand
	helpOperation := func(conf *config) error {
		fmt.Printf("\nWelcome to the Pokedex!\nUsage:\n\n")
		for _, command := range commands {
			fmt.Printf("%s: %s\n", command.name, command.description)
		}
		fmt.Println()
		return nil
	}

	exitOperation := func(conf *config) error {
		fmt.Println("Exiting the Pokedex...")
		defer os.Exit(0)
		return nil
	}

	errorOperation := func(conf *config) error {
		return errors.New("error operation")
	}

	mapOperation := func(conf *config) error {
		if conf.nextUrl == "" {
			fmt.Println("You are at the last page. No more locations. Choose *mapb* command")
			return nil
		}
		var (
			body []byte
			ok   bool
			err  error
		)

		if body, ok = cache.Get(conf.nextUrl); !ok {
			body, err = pokeRequest(conf.nextUrl)
			if err != nil {
				return err
			}
			cache.Add(conf.previousUrl, body)
		}
		curLocations := pokeRespLocation{}
		err = json.Unmarshal(body, &curLocations)
		if err != nil {
			return err
		}
		var nextUrl, prevUrl string
		if curLocations.Next != nil {
			nextUrl = *(curLocations.Next)
		}
		if curLocations.Previous != nil {
			prevUrl = *(curLocations.Previous)
		}
		conf.nextUrl, conf.previousUrl = nextUrl, prevUrl
		for _, location := range curLocations.Results {
			fmt.Println(location.Name)
		}
		return nil
	}

	mapbOperation := func(conf *config) error {
		if conf.previousUrl == "" {
			fmt.Println("You are at the first page or didn't choose *map* command. No previous locations. Choose *map* command.")
			return nil
		}
		var (
			body []byte
			ok   bool
			err  error
		)

		if body, ok = cache.Get(conf.previousUrl); !ok {
			body, err = pokeRequest(conf.previousUrl)
			if err != nil {
				return err
			}
			cache.Add(conf.previousUrl, body)
		}
		curLocations := pokeRespLocation{}
		err = json.Unmarshal(body, &curLocations)
		if err != nil {
			return err
		}
		var nextUrl, prevUrl string
		if curLocations.Next != nil {
			nextUrl = *(curLocations.Next)
		}
		if curLocations.Previous != nil {
			prevUrl = *(curLocations.Previous)
		}
		conf.nextUrl, conf.previousUrl = nextUrl, prevUrl
		for _, location := range curLocations.Results {
			fmt.Println(location.Name)
		}
		return nil
	}

	commands = map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			operation:   helpOperation,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			operation:   exitOperation,
		},
		"error": {
			name:        "error",
			description: "Error test",
			operation:   errorOperation,
		},
		"map": {
			name:        "map",
			description: "Displays the names of the 20 next location areas in the Pokemon world",
			operation:   mapOperation,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the names of the 20 previous location areas in the Pokemon world",
			operation:   mapbOperation,
		},
	}

	conf := config{
		previousUrl: "",
		nextUrl:     "https://pokeapi.co/api/v2/location?offset=0&limit=20",
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("Pokedex > ")
		scanner.Scan()
		err := scanner.Err()
		if err != nil {
			fmt.Println(err.Error())
		}
		splitline := strings.Fields(strings.ToLower(scanner.Text()))
		if len(splitline) == 0 {
			continue
		}
		command, ok := commands[(splitline[0])]
		if !ok {
			fmt.Printf("\nWrong command. Try again.\n\n")
			continue
		}
		err = command.operation(&conf)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}
