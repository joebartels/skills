package reportview

import "encoding/json"

type Formatter struct{}

func New() Formatter { return Formatter{} }
func (Formatter) Format(items []string) ([]byte, error) {
	if len(items) > 10 {
		items = items[:10]
	}
	return json.Marshal(struct {
		Items []string `json:"items"`
	}{items})
}
