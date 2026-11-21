package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"goflow/internal/model"
)

var apiClient = &http.Client{
	Timeout: 5 * time.Second,
}

func fetchSnapshot(ctx context.Context) (model.Snapshot, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"/api/state",
		nil,
	)
	if err != nil {
		return model.Snapshot{}, err
	}

	resp, err := apiClient.Do(req)

	if err != nil {
		return model.Snapshot{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.Snapshot{}, fmt.Errorf(
			"сервер вернул HTTP %d", resp.StatusCode,
		)
	}
	var snapshot model.Snapshot

	err = json.NewDecoder(resp.Body).Decode(&snapshot)

	if err != nil {
		return model.Snapshot{}, err
	}

	return snapshot, nil
}

func sendCommand(
	ctx context.Context,
	path string,
	payload any,
) (model.Snapshot, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return model.Snapshot{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		path,
		bytes.NewReader(data),
	)
	if err != nil {
		return model.Snapshot{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := apiClient.Do(req)
	if err != nil {
		return model.Snapshot{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var serverError struct {
			Error string `json:"error"`
		}

		err := json.NewDecoder(resp.Body).Decode(&serverError)
		if err != nil || serverError.Error == "" {
			return model.Snapshot{}, fmt.Errorf(
				"сервер вернул HTTP %d", resp.StatusCode,
			)
		}

		return model.Snapshot{}, fmt.Errorf(
			"HTTP %d: %s", resp.StatusCode, serverError.Error,
		)
	}

	var snapshot model.Snapshot

	err = json.NewDecoder(resp.Body).Decode(&snapshot)
	if err != nil {
		return model.Snapshot{}, err
	}

	return snapshot, nil
}
