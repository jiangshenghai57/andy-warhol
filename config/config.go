package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// typeAssertion tries to convert an interface{} to a proper Go type and returns the string representation.
// You can modify it to return the actual type instead of a string if needed.
func typeAssertion(val interface{}) string {
	switch v := val.(type) {
	case nil:
		return ""
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case float64:
		// JSON numbers are decoded as float64
		// Optionally, you can check if it's an integer value
		if v == float64(int(v)) {
			return fmt.Sprintf("%d", int(v))
		}
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case []interface{}:
		return typeAssertion((v))
	case map[string]interface{}:
		return typeAssertion(v)
	case []string:
		return fmt.Sprintf("%v", v)
	default:
		log.Println("not sure about the type here buddy")
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

	// for key, value := range result {
	// 	result[key] = typeAssertion(value)
	// }

	return result, err
}
