package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// Set up configuration elements here
// use NewConfigElement to create them only
// Do not directly access, use Getters and setters
var (
	ConfigElementsList = []*ConfigElement{
		NewConfigElement(
			"WHITELISTED_REPOSITORIES", 
			[]string{""}, 
			true,
		),
		NewConfigElement(
			"ON_WEBHOOK_COMMAND", 
			[]string{"docker", "compose", "up", "-d"}, 
			false,
		),
		NewConfigElement(
			"WEBHOOK_URL", 
			[]string{"/webhook"}, 
			false,
		),
		NewConfigElement(
			"WEB_PORT",
			[]string{"8080"},
			false,
		),
		NewConfigElement(
			"LOG_LEVEL",
			[]string{"INFO"},
			false,
		),
	}
)


type ConfigElement struct {
	name string
	required bool
	defaultValue []string
	Params []string
}

func NewConfigElement(name string, defaultValue []string, required bool) *ConfigElement {
	return &ConfigElement{
		name: name,
		required: required,
		defaultValue: defaultValue,
	}
}

func (c ConfigElement) GetName() string {
	return c.name
}

func (c ConfigElement) IsRequired() bool {
	return c.required
}

func (c ConfigElement) GetDefaultValue() []string {
	return c.defaultValue
}

func (c ConfigElement) GetParams() []string {
	return c.Params
}

func (c *ConfigElement) SetParams(params []string) {
	c.Params = params
}

// Returns the first index in params
func (c *ConfigElement) First() string {
	return c.Params[0]
}

// Returns everything after the first index in params
// OR it returns the first index if its the only one
func (c *ConfigElement) After() []string {
	if len(c.Params) > 1 {
		return c.Params[1:]
	} else {
		return c.Params
	}
}

// Environment variable MUST be set or this will stop execution
func RequiredEnv(key string) string {
	v, e := os.LookupEnv(key)
	if !e {
		log.Fatalf("Required environment variale %s not found. Please check the Readme", key)
	}
	return v
}

//Key = Environment variable you are looking for
//Fallback = Default value
func GetEnvWithDefault(key, fallback string) string {
	v, e := os.LookupEnv(key)
	if !e {
		return fallback
	}
	return v
}

// Parse .env in this directory
func ParseEnvFile() {

	envMap := make(map[string][]string)

	file, err := os.OpenFile(".env", os.O_RDONLY, 0666)
	if err != nil {
		log.Println("Error opening .env file. Is it an executable?")
		log.Println("Falling back to checking local environment...")
		ParseEnv()
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		beforeString, afterString, e := strings.Cut(scanner.Text(), "=")
		if !e {
			panic("Borked")
		}
		// We do not need any "" around the string any longer
		fixupEnv := strings.Trim(afterString, "\"")
		// Then we have to split it again by space for the params variable
		splitEnv := strings.Split(fixupEnv, " ")
		envMap[beforeString] = splitEnv
	}

	var params []string
	var e bool
	for _, configelement := range ConfigElementsList {
		params, e = envMap[configelement.GetName()]
		if configelement.IsRequired() {
			if !e {
				worldParams := RequiredEnv(configelement.GetName())
				params = strings.Split(worldParams, " ")
			}
		} else {
			if !e {
				params = configelement.GetDefaultValue()
			}
		}
		configelement.SetParams(params)
	}
}

// Fallback, parse environment variables in the environment
func ParseEnv() {

	var params []string
	var paramsString string
	for _, configelement := range ConfigElementsList {
		if configelement.IsRequired() {
			paramsString = RequiredEnv(configelement.GetName())
		} else {
			paramsString = GetEnvWithDefault(configelement.GetName(), strings.Join(configelement.GetDefaultValue(), " "))
		}
		params = strings.Split(paramsString, " ")
		configelement.SetParams(params)
	}
}

// Returns the config element to be passed to the webhandler
func GetConfig() map[string]*ConfigElement {
	// Check for .env existence
	envFilePath := ".env"
	if _, err := os.Stat(envFilePath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println(".env file not found. Defaulting to local environment vars")
		} else {
			fmt.Printf("Error checking .env file: %v\n", err)
		}
		// Not found? Use default parsing
		ParseEnv()
	} else {
		fmt.Println(".env file found.")
		// Env File parsing
		ParseEnvFile()
	}
	// put config elements in to a map for easy access
	ConfigObject := make(map[string]*ConfigElement)

	for _, config := range ConfigElementsList {
		ConfigObject[config.GetName()] = config
		//TODO: Proper logging
		//fmt.Printf("%s = %v\n", config.GetName(), config)
	}
	return ConfigObject
}