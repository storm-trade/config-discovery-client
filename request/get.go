package request

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func Get[T any](uri string) (T, error) {
	var result T

	resp, err := http.Get(uri)
	if err != nil {
		return result, fmt.Errorf("failed to fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("failed to read response body: %w", err)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return result, nil
}
