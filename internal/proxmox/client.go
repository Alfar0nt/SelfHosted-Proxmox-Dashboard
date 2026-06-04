package proxmox

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type NodeStatus struct {
	CPU    float64 `json:"cpu"`
	Memory struct {
		Total int64 `json:"total"`
		Used  int64 `json:"used"`
		Free  int64 `json:"free"`
	} `json:"memory"`
	Disk struct {
		Total int64 `json:"total"`
		Used  int64 `json:"used"`
	} `json:"rootfs"`
	Uptime int64 `json:"uptime"`
}

type Client struct {
	url         string
	nodeName    string
	tokenID     string
	tokenSecret string
	mock        bool
	httpClient  *http.Client
}

func NewClient(url, nodeName, tokenID, tokenSecret string, mock bool) *Client {
	// Proxmox often uses self-signed certificates
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &Client{
		url:         url,
		nodeName:    nodeName,
		tokenID:     tokenID,
		tokenSecret: tokenSecret,
		mock:        mock,
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   5 * time.Second,
		},
	}
}

func (c *Client) GetNodeStatus() (NodeStatus, error) {
	if c.mock {
		return getMockStatus(), nil
	}

	endpoint := fmt.Sprintf("%s/nodes/%s/status", c.url, url.PathEscape(c.nodeName))
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return NodeStatus{}, err
	}

	// Set Proxmox API Token header
	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", c.tokenID, c.tokenSecret))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NodeStatus{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return NodeStatus{}, fmt.Errorf("proxmox API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Data NodeStatus `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return NodeStatus{}, err
	}

	return result.Data, nil
}

func getMockStatus() NodeStatus {
	// Generate some random varying stats for testing the UI
	var status NodeStatus
	status.CPU = 0.15 + rand.Float64()*(0.65-0.15) // Random CPU between 15% and 65%

	status.Memory.Total = 16 * 1024 * 1024 * 1024                                         // 16GB
	status.Memory.Used = int64(float64(status.Memory.Total) * (0.4 + rand.Float64()*0.2)) // 40-60% used
	status.Memory.Free = status.Memory.Total - status.Memory.Used

	status.Disk.Total = 256 * 1024 * 1024 * 1024 // 256GB
	status.Disk.Used = 120 * 1024 * 1024 * 1024  // ~120GB used

	status.Uptime = 3600 * 24 * 7 // 7 days

	return status
}
