package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Pastry struct {
	Name        string
	Description string
	Size        string
	Price       float32
	Status      string
}

type PastryAPI interface {
	GetPastry(name string) (*Pastry, error)
	ListPastries(size string) (*[]Pastry, error)
}

type pastryAPIClient struct {
	baseURL string
}

func NewPastryAPIClient(baseURL string) PastryAPI {
	return &pastryAPIClient{
		baseURL: baseURL,
	}
}

func (c *pastryAPIClient) GetPastry(name string) (*Pastry, error) {
	url := c.baseURL + "/pastries/" + name

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var pastry Pastry
	err = json.NewDecoder(resp.Body).Decode(&pastry)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &pastry, nil
}

func (c *pastryAPIClient) ListPastries(size string) (*[]Pastry, error) {
	url := c.baseURL + "/pastries?size=" + size

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var pastries []Pastry
	err = json.NewDecoder(resp.Body).Decode(&pastries)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &pastries, nil
}
