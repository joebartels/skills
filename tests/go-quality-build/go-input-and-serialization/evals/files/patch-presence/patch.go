package configpatch

import (
	"encoding/json"
	"errors"
)

type Config struct {
	Enabled bool              `json:"enabled"`
	Retries int               `json:"retries"`
	Label   *string           `json:"label"`
	Labels  map[string]string `json:"labels"`
}

// Apply updates the supplied configuration.
func Apply(dst *Config, data []byte) error {
	if err := json.Unmarshal(data, dst); err != nil {
		return err
	}
	if dst.Retries < 0 || dst.Retries > 10 {
		return errors.New("retries must be between 0 and 10")
	}
	return nil
}
