package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

func TestReadConfig_Local(t *testing.T) {
	// Prepare a temporary config file in the local directory
	configData := map[string]interface{}{
		"foo":   "bar",
		"count": 42,
	}
	configBytes, _ := json.Marshal(configData)
	configFile := "./config.json"
	defer os.Remove(configFile) // clean up after test

	if err := ioutil.WriteFile(configFile, configBytes, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	// Ensure OCP_EN is unset
	os.Unsetenv("OCP_EN")
	os.Unsetenv("CONFIG_PATH")

	result, err := ReadConfig()
	if err != nil {
		t.Errorf("ReadConfig returned error: %v", err)
	}
	if result["foo"] != "bar" || int(result["count"].(float64)) != 42 {
		t.Errorf("Config contents incorrect: got %v", result)
	}
}

func TestReadConfig_Kubernetes(t *testing.T) {
	// Prepare a temporary config file in a custom path
	configData := map[string]interface{}{
		"env": "k8s",
	}
	configBytes, _ := json.Marshal(configData)
	tmpDir := os.TempDir() + "/andy-warhol/"
	os.MkdirAll(tmpDir, 0755)
	configFile := tmpDir + "config.json"
	defer os.Remove(configFile)
	defer os.Remove(tmpDir)

	if err := ioutil.WriteFile(configFile, configBytes, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment variables
	os.Setenv("OCP_EN", "true")
	os.Setenv("CONFIG_PATH", tmpDir)

	result, err := ReadConfig()
	if err != nil {
		t.Errorf("ReadConfig returned error: %v", err)
	}
	if result["env"] != "k8s" {
		t.Errorf("Config contents incorrect: got %v", result)
	}
}
