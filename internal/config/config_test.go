package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, dir string, data map[string]interface{}) string {
	configBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	configFile := filepath.Join(dir, "config.json")
	if err := os.WriteFile(configFile, configBytes, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	return configFile
}

func TestReadConfig_Local(t *testing.T) {
	// Prepare a temporary config file in the local directory
	configData := map[string]interface{}{
		"foo":   "bar",
		"count": 42,
	}
	configFile := writeTempConfig(t, ".", configData)
	defer os.Remove(configFile)

	os.Unsetenv("OCP_ENV")
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
	cwd, _ := os.Getwd()
	parentDir := filepath.Dir(cwd)
	configFile := writeTempConfig(t, parentDir, configData)
	defer os.Remove(configFile)

	os.Setenv("OCP_ENV", "true")
	os.Setenv("CONFIG_PATH", parentDir+string(os.PathSeparator))

	result, err := ReadConfig()
	if err != nil {
		t.Errorf("ReadConfig returned error: %v", err)
	}
	if result["env"] != "k8s" {
		t.Errorf("Config contents incorrect: got %v", result)
	}
}

func TestConvertTypes(t *testing.T) {
	input := map[string]interface{}{
		"intVal":   int(10),
		"floatVal": float64(10.5),
		"strVal":   "hello",
		"arrVal":   []interface{}{float64(1), "two", float64(3.0)},
		"mapVal": map[string]interface{}{
			"nestedInt": float64(7),
			"nestedStr": "world",
		},
	}
	expected := map[string]interface{}{
		"intVal":   10,
		"floatVal": 10.5,
		"strVal":   "hello",
		"arrVal":   []interface{}{float64(1), "two", float64(3.0)},
		"mapVal": map[string]interface{}{
			"nestedInt": float64(7),
			"nestedStr": "world",
		},
	}
	result := convertTypes(input).(map[string]interface{})

	if result["intVal"] != expected["intVal"] {
		t.Errorf("Expected intVal %v, got %v", expected["intVal"], result["intVal"])
	}
	if result["floatVal"] != expected["floatVal"] {
		t.Errorf("Expected floatVal %v, got %v", expected["floatVal"], result["floatVal"])
	}
	if result["strVal"] != expected["strVal"] {
		t.Errorf("Expected strVal %v, got %v", expected["strVal"], result["strVal"])
	}
	arr := result["arrVal"].([]interface{})
	expArr := expected["arrVal"].([]interface{})
	for i := range arr {
		if arr[i] != expArr[i] {
			t.Errorf("Expected arrVal[%d] %v, got %v", i, expArr[i], arr[i])
		}
	}
	nested := result["mapVal"].(map[string]interface{})
	expNested := expected["mapVal"].(map[string]interface{})
	for k := range nested {
		if nested[k] != expNested[k] {
			t.Errorf("Expected mapVal[%s] %v, got %v", k, expNested[k], nested[k])
		}
	}
}
