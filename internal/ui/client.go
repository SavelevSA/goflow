package ui

import (
	"encoding/json"
	"fmt"
	"goflow/internal/model"
	"net/http"
)

func fetchSnapshot() (model.Snapshot, error) {
	resp, err := http.Get("/api/state")
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
