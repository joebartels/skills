package showcfg

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	Endpoint string `json:"endpoint"`
	Mode     string `json:"mode"`
}

// LoadFromEnv reads configuration from the current process environment.
func LoadFromEnv() (Config, error) {
	endpoint, present := os.LookupEnv("INDEXER_ENDPOINT")
	if !present || endpoint == "" {
		endpoint = "http://127.0.0.1:9090"
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return Config{}, fmt.Errorf("invalid INDEXER_ENDPOINT %q", endpoint)
	}
	mode, present := os.LookupEnv("INDEXER_MODE")
	if !present {
		mode = "read"
	}
	if mode != "read" && mode != "write" {
		return Config{}, fmt.Errorf("invalid INDEXER_MODE %q", mode)
	}
	return Config{Endpoint: endpoint, Mode: mode}, nil
}
