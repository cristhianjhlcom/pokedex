package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	cache "github.com/cristhianjhlcom/pokedex/internal"
)

const baseURL = "https://pokeapi.co/api/v2"

func main() {
	config := Config{
		pokeAPIClient: NewClient(time.Hour),
	}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("pokedex > ")
		scanner.Scan()
		text := scanner.Text()
		cleaned := cleanInput(text)
		if len(cleaned) == 0 {
			continue
		}
		commandName := cleaned[0]
		availableCommands := getCommands()
		command, ok := availableCommands[commandName]
		if !ok {
			fmt.Println("invalid command")
			continue
		}
		err := command.callback(&config)
		if err != nil {
			fmt.Println(err)
		}
	}
}

type Config struct {
	pokeAPIClient           Client
	nextLocationAreaURL     *string
	previousLocationAreaURL *string
}

type CLICommand struct {
	name        string
	description string
	callback    func(*Config) error
}

type Client struct {
	cache      cache.Cache
	httpClient http.Client
}

type LocationAreaResponse struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

func NewClient(cacheInterval time.Duration) Client {
	return Client{
		cache: cache.NewCache(cacheInterval),
		httpClient: http.Client{
			Timeout: time.Minute,
		},
	}
}

func (c *Client) ListLocationAreas(pageURL *string) (LocationAreaResponse, error) {
	endpoint := "/location/"
	fullURL := baseURL + endpoint
	if pageURL != nil {
		fullURL = *pageURL
	}
	data, ok := c.cache.Get(fullURL)
	if ok {
		// cache hit.
		locationAreasResponse := LocationAreaResponse{}
		err := json.Unmarshal(data, &locationAreasResponse)
		if err != nil {
			return LocationAreaResponse{}, err
		}
		return locationAreasResponse, nil
	}
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return LocationAreaResponse{}, err
	}
	response, err := c.httpClient.Do(req)
	if err != nil {
		return LocationAreaResponse{}, err
	}
	defer response.Body.Close()
	if response.StatusCode > 399 {
		return LocationAreaResponse{}, fmt.Errorf("bad status code: %v", response.StatusCode)
	}
	data, err = io.ReadAll(response.Body)
	if err != nil {
		return LocationAreaResponse{}, err
	}
	locationAreasResponse := LocationAreaResponse{}
	err = json.Unmarshal(data, &locationAreasResponse)
	if err != nil {
		return LocationAreaResponse{}, err
	}
	c.cache.Add(fullURL, data)
	return locationAreasResponse, nil
}

func getCommands() map[string]CLICommand {
	return map[string]CLICommand{
		"help": {
			name:        "help",
			description: "Prints the help menu",
			callback:    callbackHelp,
		},
		"map": {
			name:        "map",
			description: "Lists some locations areas",
			callback:    callbackMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Lists the previous page of locations areas",
			callback:    callbackMapb,
		},
		"exit": {
			name:        "exit",
			description: "Turns off the pokedex",
			callback:    callbackExit,
		},
	}
}

func callbackHelp(config *Config) error {
	fmt.Println("Welcome to the Pokedex help menu!")
	fmt.Println("Here are you available commands: ")
	availableCommands := getCommands()
	for _, cmd := range availableCommands {
		fmt.Printf(" - %s: %s\n", cmd.name, cmd.description)
	}
	fmt.Println("")
	return nil
}

func callbackExit(config *Config) error {
	os.Exit(0)
	return nil
}

func callbackMap(config *Config) error {
	response, err := config.pokeAPIClient.ListLocationAreas(config.nextLocationAreaURL)
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Println("Location areas")
	for _, area := range response.Results {
		fmt.Printf(" - %s\n", area.Name)
	}
	config.nextLocationAreaURL = response.Next
	config.previousLocationAreaURL = response.Previous
	return nil
}

func callbackMapb(config *Config) error {
	if config.previousLocationAreaURL == nil {
		return errors.New("You are on the first page")
	}
	response, err := config.pokeAPIClient.ListLocationAreas(config.previousLocationAreaURL)
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Println("Location areas")
	for _, area := range response.Results {
		fmt.Printf(" - %s\n", area.Name)
	}
	config.nextLocationAreaURL = response.Next
	config.previousLocationAreaURL = response.Previous
	return nil
}

func cleanInput(str string) []string {
	lowered := strings.ToLower(str)
	words := strings.Fields(lowered)
	return words
}
