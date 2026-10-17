package bundlecfg

import (
	"encoding/json"
	"fmt"
	"io"
)

const MaxDocumentBytes = 512
const MaxServices = 3

type Service struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

type Bundle struct {
	Services []Service `json:"services"`
}

// Load reads a service bundle and replaces the active configuration.
func Load(r io.Reader, replace func(Bundle) error) error {
	var bundle Bundle
	if err := json.NewDecoder(r).Decode(&bundle); err != nil {
		return err
	}
	for _, service := range bundle.Services {
		if service.Name == "" || service.Port < 1 || service.Port > 65535 {
			return fmt.Errorf("invalid service %q", service.Name)
		}
	}
	return replace(bundle)
}
