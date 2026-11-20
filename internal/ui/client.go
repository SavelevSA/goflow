package ui

import (
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
