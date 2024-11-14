// Copyright The Microcks Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Pastry struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Size        string  `json:"size"`
	Price       float32 `json:"price"`
	Status      string  `json:"status"`
}

type PastryAPI interface {
	GetPastry(name string) (Pastry, error)
	ListPastries(size string) ([]Pastry, error)
}

type pastryAPIClient struct {
	baseURL string
}

func NewPastryAPIClient(baseURL string) PastryAPI {
	return &pastryAPIClient{
		baseURL: baseURL,
	}
}

func (c *pastryAPIClient) GetPastry(name string) (Pastry, error) {
	url := c.baseURL + "/pastries/" + name

	resp, err := http.Get(url)
	if err != nil {
		return Pastry{}, fmt.Errorf("failed to make GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Pastry{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var pastry Pastry
	err = json.NewDecoder(resp.Body).Decode(&pastry)
	if err != nil {
		return Pastry{}, fmt.Errorf("failed to decode response body: %w", err)
	}

	return pastry, nil
}

func (c *pastryAPIClient) ListPastries(size string) ([]Pastry, error) {
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

	return pastries, nil
}
