package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// typeAssertion tries to convert an interface{} to a proper Go type and returns the string representation.
// You can modify it to return the actual type instead of a string if needed.
func convertTypes(val interface{}) interface{} {
	switch v := val.(type) {
	case map[string]interface{}:
		m := make(map[string]interface{})
		for key, value := range v {
			m[key] = convertTypes(value)
		}
		return m
	case []interface{}:
		arr := make([]interface{}, len(v))
		for i, elem := range v {
			arr[i] = convertTypes(elem)
		}
		return arr
	case float64:
		return v
	case int:
		return v
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func ReadConfig() (map[string]interface{}, error) {
	OCP_ENV := os.Getenv("OCP_ENV")
	CONFIG_PATH := os.Getenv("CONFIG_PATH")

	var config_path_file = ""

	if OCP_ENV == "" {
		config_path_file = "./config.json"
	} else {
		config_path_file = CONFIG_PATH + "config.json"
	}

	log.Println("Reading in config from:", config_path_file)
	file, err := os.Open(config_path_file)

	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Decode into a map
	var result map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&result); err != nil {
		panic(err)
	}

	result = convertTypes(result).(map[string]interface{})

	return result, err
}
