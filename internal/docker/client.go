package docker

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"
)

type Container struct {
	Id     string   `json:"Id"`
	Names  []string `json:"Names"`
	State  string   `json:"State"`
	Status string   `json:"Status"`
}

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(endpoint string) *Client {
	var tr *http.Transport
	baseURL := "http://localhost" // default for unix socket

	if strings.HasPrefix(endpoint, "unix://") {
		path := strings.TrimPrefix(endpoint, "unix://")
		tr = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", path)
			},
		}
	} else if strings.HasPrefix(endpoint, "tcp://") {
		// Convert tcp:// to http:// for the HTTP client
		baseURL = strings.Replace(endpoint, "tcp://", "http://", 1)
		tr = &http.Transport{}
	} else if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		baseURL = endpoint
		tr = &http.Transport{}
	} else {
		// Fallback assuming it's a raw socket path
		path := endpoint
		tr = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", path)
			},
		}
	}

	return &Client{
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   5 * time.Second,
		},
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

// GetContainers fetches the list of containers.
func (c *Client) GetContainers() ([]Container, error) {
	// Docker API endpoint for listing all containers
	url := c.baseURL + "/containers/json?all=1"
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var containers []Container
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, err
	}

	return containers, nil
}
