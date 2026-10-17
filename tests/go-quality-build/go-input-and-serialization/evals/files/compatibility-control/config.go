package config

import "encoding/json"

// Config is the persisted client configuration.
type Config struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// Decode reads the compatible JSON v1 configuration format.
func Decode(data []byte) (Config, error) {
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Encode preserves the distinction between nil and empty Tags.
func Encode(cfg Config) ([]byte, error) {
	return json.Marshal(cfg)
}
