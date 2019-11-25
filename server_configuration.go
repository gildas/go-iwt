package iwt

import (
	"fmt"
	"net/http"

	"github.com/gildas/go-request"
)

// ServerConfiguration contains information about a PureConnect server
type ServerConfiguration struct {
	Version      int                 `json:"cfgVer"`
	Capabilities map[string][]string `json:"capabilities"`
}

// GetServerConfiguration fetches the configuration of the PureConnect server
func (client *Client) GetServerConfiguration() (*ServerConfiguration, error) {
	results := []struct {
		Config ServerConfiguration `json:"serverConfiguration"`
	}{}
	_, err := client.get("/serverConfiguration", &results)
	if reqerr, ok := err.(request.Error); ok && reqerr.StatusCode == http.StatusServiceUnavailable {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("Failed to query")
	}

	return &results[0].Config, nil
}
