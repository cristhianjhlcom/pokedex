package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
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
		caughtPokemon: make(map[string]Pokemon),
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
		args := []string{}
		if len(cleaned) > 1 {
			args = cleaned[1:]
		}
		availableCommands := getCommands()
		command, ok := availableCommands[commandName]
		if !ok {
			fmt.Println("invalid command")
			continue
		}
		err := command.callback(&config, args...)
		if err != nil {
			fmt.Println(err)
		}
	}
}

type Config struct {
	pokeAPIClient           Client
	nextLocationAreaURL     *string
	previousLocationAreaURL *string
	caughtPokemon           map[string]Pokemon
}

type CLICommand struct {
	name        string
	description string
	callback    func(*Config, ...string) error
}

type Pokemon struct {
	Abilities []struct {
		Ability struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"ability"`
		IsHidden bool `json:"is_hidden"`
		Slot     int  `json:"slot"`
	} `json:"abilities"`
	BaseExperience int `json:"base_experience"`
	Cries          struct {
		Latest string `json:"latest"`
		Legacy string `json:"legacy"`
	} `json:"cries"`
	Forms []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"forms"`
	GameIndices []struct {
		GameIndex int `json:"game_index"`
		Version   struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"version"`
	} `json:"game_indices"`
	Height                 int           `json:"height"`
	HeldItems              []interface{} `json:"held_items"`
	ID                     int           `json:"id"`
	IsDefault              bool          `json:"is_default"`
	LocationAreaEncounters string        `json:"location_area_encounters"`
	Moves                  []struct {
		Move struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"move"`
		VersionGroupDetails []struct {
			LevelLearnedAt  int `json:"level_learned_at"`
			MoveLearnMethod struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"move_learn_method"`
			VersionGroup struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version_group"`
		} `json:"version_group_details"`
	} `json:"moves"`
	Name          string        `json:"name"`
	Order         int           `json:"order"`
	PastAbilities []interface{} `json:"past_abilities"`
	PastTypes     []interface{} `json:"past_types"`
	Species       struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"species"`
	Sprites struct {
		BackDefault      string      `json:"back_default"`
		BackFemale       interface{} `json:"back_female"`
		BackShiny        string      `json:"back_shiny"`
		BackShinyFemale  interface{} `json:"back_shiny_female"`
		FrontDefault     string      `json:"front_default"`
		FrontFemale      interface{} `json:"front_female"`
		FrontShiny       string      `json:"front_shiny"`
		FrontShinyFemale interface{} `json:"front_shiny_female"`
		Other            struct {
			DreamWorld struct {
				FrontDefault string      `json:"front_default"`
				FrontFemale  interface{} `json:"front_female"`
			} `json:"dream_world"`
			Home struct {
				FrontDefault     string      `json:"front_default"`
				FrontFemale      interface{} `json:"front_female"`
				FrontShiny       string      `json:"front_shiny"`
				FrontShinyFemale interface{} `json:"front_shiny_female"`
			} `json:"home"`
			OfficialArtwork struct {
				FrontDefault string `json:"front_default"`
				FrontShiny   string `json:"front_shiny"`
			} `json:"official-artwork"`
			Showdown struct {
				BackDefault      string      `json:"back_default"`
				BackFemale       interface{} `json:"back_female"`
				BackShiny        string      `json:"back_shiny"`
				BackShinyFemale  interface{} `json:"back_shiny_female"`
				FrontDefault     string      `json:"front_default"`
				FrontFemale      interface{} `json:"front_female"`
				FrontShiny       string      `json:"front_shiny"`
				FrontShinyFemale interface{} `json:"front_shiny_female"`
			} `json:"showdown"`
		} `json:"other"`
		Versions struct {
			GenerationI struct {
				RedBlue struct {
					BackDefault      string `json:"back_default"`
					BackGray         string `json:"back_gray"`
					BackTransparent  string `json:"back_transparent"`
					FrontDefault     string `json:"front_default"`
					FrontGray        string `json:"front_gray"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"red-blue"`
				Yellow struct {
					BackDefault      string `json:"back_default"`
					BackGray         string `json:"back_gray"`
					BackTransparent  string `json:"back_transparent"`
					FrontDefault     string `json:"front_default"`
					FrontGray        string `json:"front_gray"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"yellow"`
			} `json:"generation-i"`
			GenerationIi struct {
				Crystal struct {
					BackDefault           string `json:"back_default"`
					BackShiny             string `json:"back_shiny"`
					BackShinyTransparent  string `json:"back_shiny_transparent"`
					BackTransparent       string `json:"back_transparent"`
					FrontDefault          string `json:"front_default"`
					FrontShiny            string `json:"front_shiny"`
					FrontShinyTransparent string `json:"front_shiny_transparent"`
					FrontTransparent      string `json:"front_transparent"`
				} `json:"crystal"`
				Gold struct {
					BackDefault      string `json:"back_default"`
					BackShiny        string `json:"back_shiny"`
					FrontDefault     string `json:"front_default"`
					FrontShiny       string `json:"front_shiny"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"gold"`
				Silver struct {
					BackDefault      string `json:"back_default"`
					BackShiny        string `json:"back_shiny"`
					FrontDefault     string `json:"front_default"`
					FrontShiny       string `json:"front_shiny"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"silver"`
			} `json:"generation-ii"`
			GenerationIii struct {
				Emerald struct {
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"emerald"`
				FireredLeafgreen struct {
					BackDefault  string `json:"back_default"`
					BackShiny    string `json:"back_shiny"`
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"firered-leafgreen"`
				RubySapphire struct {
					BackDefault  string `json:"back_default"`
					BackShiny    string `json:"back_shiny"`
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"ruby-sapphire"`
			} `json:"generation-iii"`
			GenerationIv struct {
				DiamondPearl struct {
					BackDefault      string      `json:"back_default"`
					BackFemale       interface{} `json:"back_female"`
					BackShiny        string      `json:"back_shiny"`
					BackShinyFemale  interface{} `json:"back_shiny_female"`
					FrontDefault     string      `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       string      `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"diamond-pearl"`
				HeartgoldSoulsilver struct {
					BackDefault      string      `json:"back_default"`
					BackFemale       interface{} `json:"back_female"`
					BackShiny        string      `json:"back_shiny"`
					BackShinyFemale  interface{} `json:"back_shiny_female"`
					FrontDefault     string      `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       string      `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"heartgold-soulsilver"`
				Platinum struct {
					BackDefault      string      `json:"back_default"`
					BackFemale       interface{} `json:"back_female"`
					BackShiny        string      `json:"back_shiny"`
					BackShinyFemale  interface{} `json:"back_shiny_female"`
					FrontDefault     string      `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       string      `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"platinum"`
			} `json:"generation-iv"`
			GenerationV struct {
				BlackWhite struct {
					Animated struct {
						BackDefault      string      `json:"back_default"`
						BackFemale       interface{} `json:"back_female"`
						BackShiny        string      `json:"back_shiny"`
						BackShinyFemale  interface{} `json:"back_shiny_female"`
						FrontDefault     string      `json:"front_default"`
						FrontFemale      interface{} `json:"front_female"`
						FrontShiny       string      `json:"front_shiny"`
						FrontShinyFemale interface{} `json:"front_shiny_female"`
					} `json:"animated"`
					BackDefault      string      `json:"back_default"`
					BackFemale       interface{} `json:"back_female"`
					BackShiny        string      `json:"back_shiny"`
					BackShinyFemale  interface{} `json:"back_shiny_female"`
					FrontDefault     string      `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       string      `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"black-white"`
			} `json:"generation-v"`
			GenerationVi struct {
				OmegarubyAlphasapphire struct {
					FrontDefault     string      `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       string      `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"omegaruby-alphasapphire"`
				XY struct {
					FrontDefault     string      `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       string      `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"x-y"`
			} `json:"generation-vi"`
			GenerationVii struct {
				Icons struct {
					FrontDefault string      `json:"front_default"`
					FrontFemale  interface{} `json:"front_female"`
				} `json:"icons"`
				UltraSunUltraMoon struct {
					FrontDefault     string      `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       string      `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"ultra-sun-ultra-moon"`
			} `json:"generation-vii"`
			GenerationViii struct {
				Icons struct {
					FrontDefault string      `json:"front_default"`
					FrontFemale  interface{} `json:"front_female"`
				} `json:"icons"`
			} `json:"generation-viii"`
		} `json:"versions"`
	} `json:"sprites"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	Weight int `json:"weight"`
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

type LocationArea struct {
	Areas []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"areas"`
	GameIndices []struct {
		GameIndex  int `json:"game_index"`
		Generation struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"generation"`
	} `json:"game_indices"`
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Names []struct {
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
		Name string `json:"name"`
	} `json:"names"`
	Region struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"region"`
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

func (c *Client) GetLocationArea(locationAreaName string) (LocationArea, error) {
	endpoint := "/location/" + locationAreaName
	fullURL := baseURL + endpoint
	data, ok := c.cache.Get(fullURL)
	if ok {
		// cache hit.
		locationArea := LocationArea{}
		err := json.Unmarshal(data, &locationArea)
		if err != nil {
			return LocationArea{}, err
		}
		return locationArea, nil
	}
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return LocationArea{}, err
	}
	response, err := c.httpClient.Do(req)
	if err != nil {
		return LocationArea{}, err
	}
	defer response.Body.Close()
	if response.StatusCode > 399 {
		return LocationArea{}, fmt.Errorf("bad status code: %v", response.StatusCode)
	}
	data, err = io.ReadAll(response.Body)
	if err != nil {
		return LocationArea{}, err
	}
	locationArea := LocationArea{}
	err = json.Unmarshal(data, &locationArea)
	if err != nil {
		return LocationArea{}, err
	}
	c.cache.Add(fullURL, data)
	return locationArea, nil
}

func (c *Client) GetPokemon(name string) (Pokemon, error) {
	endpoint := "/pokemon/" + name
	fullURL := baseURL + endpoint
	data, ok := c.cache.Get(fullURL)
	if ok {
		// cache hit.
		pokemon := Pokemon{}
		err := json.Unmarshal(data, &pokemon)
		if err != nil {
			return Pokemon{}, err
		}
		return pokemon, nil
	}
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return Pokemon{}, err
	}
	response, err := c.httpClient.Do(req)
	if err != nil {
		return Pokemon{}, err
	}
	defer response.Body.Close()
	if response.StatusCode > 399 {
		return Pokemon{}, fmt.Errorf("bad status code: %v", response.StatusCode)
	}
	data, err = io.ReadAll(response.Body)
	if err != nil {
		return Pokemon{}, err
	}
	pokemon := Pokemon{}
	err = json.Unmarshal(data, &pokemon)
	if err != nil {
		return Pokemon{}, err
	}
	c.cache.Add(fullURL, data)
	return pokemon, nil
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
		"explore": {
			name:        "explore {location_area}",
			description: "List the areas in location",
			callback:    callbackExplorer,
		},
		"catch": {
			name:        "catch {pokemon_name}",
			description: "Attempt to catch a pokemon and add it to your pokedex",
			callback:    callbackCatch,
		},
		"exit": {
			name:        "exit",
			description: "Turns off the pokedex",
			callback:    callbackExit,
		},
	}
}

func callbackHelp(config *Config, args ...string) error {
	fmt.Println("Welcome to the Pokedex help menu!")
	fmt.Println("Here are you available commands: ")
	availableCommands := getCommands()
	for _, cmd := range availableCommands {
		fmt.Printf(" - %s: %s\n", cmd.name, cmd.description)
	}
	fmt.Println("")
	return nil
}

func callbackExit(config *Config, args ...string) error {
	os.Exit(0)
	return nil
}

func callbackExplorer(config *Config, args ...string) error {
	if len(args) != 1 {
		return errors.New("No location area provided")
	}
	locationArea := args[0]
	response, err := config.pokeAPIClient.GetLocationArea(locationArea)
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Printf("Areas in %s \n", locationArea)
	for _, area := range response.Areas {
		fmt.Printf(" - %s\n", area.Name)
	}
	return nil
}

func callbackCatch(config *Config, args ...string) error {
	if len(args) != 1 {
		return errors.New("No pokemon name provided")
	}
	pokemonName := args[0]
	response, err := config.pokeAPIClient.GetPokemon(pokemonName)
	if err != nil {
		log.Fatal(err)
		return err
	}
	const threshold = 50
	randomNumber := rand.Intn(response.BaseExperience)
	fmt.Println(response.BaseExperience, randomNumber, threshold)
	if randomNumber > threshold {
		return fmt.Errorf("Failed to catch %s!\n", pokemonName)
	}
	config.caughtPokemon[pokemonName] = response
	fmt.Printf("%s was caught!\n", pokemonName)
	return nil
}

func callbackMap(config *Config, args ...string) error {
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

func callbackMapb(config *Config, args ...string) error {
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
