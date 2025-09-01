package config

import (
	"encoding/json"
	"log"
	"os"
)

func ReadConfig() (map[string]interface{}, error) {
	OCP_ENV := os.Getenv("OCP_EN")
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

	return result, err
}
