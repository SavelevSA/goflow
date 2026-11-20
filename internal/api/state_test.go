package api

import (
	"encoding/json"
	"goflow/internal/model"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetState(t *testing.T) {
	mux := http.NewServeMux()
	RegisterPreview(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/state", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("статус = %d, ожидали 200. Ответ: %s", rec.Code, rec.Body.String())
	}

	var snapshot model.Snapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("не удалось прочитать снимок: %v", err)
	}

	if snapshot.Density != 3 {
		t.Errorf("плотность = %d, ожидали 3", snapshot.Density)
	}

	if len(snapshot.Intersections) != 4 {
		t.Errorf("перекрёстки = %p, ожидали 4", snapshot.Intersections)
	}
}

func TestSetDensity(t *testing.T) {
	mux := http.NewServeMux()
	RegisterPreview(mux)

	postReq := httptest.NewRequest(http.MethodPost, "/api/density", strings.NewReader(`{"density":5}`))
	postReq.Header.Set("Content-Type", "application/json")
	postRec := httptest.NewRecorder()
	mux.ServeHTTP(postRec, postReq)

	if postRec.Code != 200 {
		t.Fatalf("статус = %d, ожидали 200. Ответ: %s", postRec.Code, postRec.Body.String())
	}

	var snapshot model.Snapshot
	if err := json.Unmarshal(postRec.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("не удалось прочитать снимок: %v", err)
	}

	if snapshot.Density != 5 {
		t.Errorf("после POST плотность = %d, ожидали 5", snapshot.Density)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/state", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("статус = %d, ожидали 200. Ответ: %s", rec.Code, rec.Body.String())
	}

	var getSnapshot model.Snapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &getSnapshot); err != nil {
		t.Fatalf("не удалось прочитать снимок после GET: %v", err)
	}

	if getSnapshot.Density != 5 {
		t.Errorf("после GET плотность = %d, ожидали 5", getSnapshot.Density)
	}
}

func TestSetPaused(t *testing.T) {
	mux := http.NewServeMux()
	RegisterPreview(mux)

	postReq := httptest.NewRequest(http.MethodPost, "/api/pause", strings.NewReader(`{"paused":true}`))
	postReq.Header.Set("Content-Type", "application/json")
	postRec := httptest.NewRecorder()
	mux.ServeHTTP(postRec, postReq)

	if postRec.Code != 200 {
		t.Fatalf("статус = %d, ожидали 200. Ответ: %s", postRec.Code, postRec.Body.String())
	}

	var snapshot model.Snapshot
	if err := json.Unmarshal(postRec.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("не удалось прочитать снимок: %v", err)
	}

	if !snapshot.Paused {
		t.Errorf("после POST с paused=true ожидали paused=true")
	}

	postReq2 := httptest.NewRequest(http.MethodPost, "/api/pause", strings.NewReader(`{"paused":false}`))
	postReq2.Header.Set("Content-Type", "application/json")
	postRec2 := httptest.NewRecorder()
	mux.ServeHTTP(postRec2, postReq2)

	if postRec2.Code != 200 {
		t.Fatalf("статус = %d, ожидали 200. Ответ: %s", postRec2.Code, postRec2.Body.String())
	}

	var snapshot2 model.Snapshot
	if err := json.Unmarshal(postRec2.Body.Bytes(), &snapshot2); err != nil {
		t.Fatalf("не удалось прочитать снимок: %v", err)
	}

	if snapshot2.Paused {
		t.Errorf("после POST с paused=false ожидали paused=false")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/state", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("статус = %d, ожидали 200. Ответ: %s", rec.Code, rec.Body.String())
	}

	var getSnapshot model.Snapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &getSnapshot); err != nil {
		t.Fatalf("не удалось прочитать снимок после GET: %v", err)
	}

	if getSnapshot.Paused {
		t.Errorf("после GET paused = %t, ожидали false", getSnapshot.Paused)
	}
}

func TestSetMode(t *testing.T) {
	mux := http.NewServeMux()
	RegisterPreview(mux)

	postReq := httptest.NewRequest(http.MethodPost, "/api/mode", strings.NewReader(`{"mode":"adaptive"}`))
	postReq.Header.Set("Content-Type", "application/json")
	postRec := httptest.NewRecorder()
	mux.ServeHTTP(postRec, postReq)

	if postRec.Code != 200 {
		t.Fatalf("статус = %d, ожидали 200. Ответ: %s", postRec.Code, postRec.Body.String())
	}

	var snapshot model.Snapshot
	if err := json.Unmarshal(postRec.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("не удалось прочитать снимок: %v", err)
	}

	if snapshot.Mode != model.ModeAdaptive {
		t.Errorf("после POST mode = %q, ожидали %q", snapshot.Mode, model.ModeAdaptive)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/state", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("статус = %d, ожидали 200. Ответ: %s", rec.Code, rec.Body.String())
	}

	var getSnapshot model.Snapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &getSnapshot); err != nil {
		t.Fatalf("не удалось прочитать снимок после GET: %v", err)
	}

	if getSnapshot.Mode != model.ModeAdaptive {
		t.Errorf("после GET mode = %s, ожидали adaptive", getSnapshot.Mode)
	}
}

func TestReset(t *testing.T) {
	mux := http.NewServeMux()
	RegisterPreview(mux)

	postReq := httptest.NewRequest(
		http.MethodPost,
		"/api/reset",
		strings.NewReader(`{"seed":0}`),
	)
	postReq.Header.Set("Content-Type", "application/json")
	postRec := httptest.NewRecorder()
	mux.ServeHTTP(postRec, postReq)

	if postRec.Code != http.StatusOK {
		t.Fatalf("POST /api/reset вернул %d, ожидали 200. Ответ: %s",
			postRec.Code, postRec.Body.String())
	}

	type resetSnapshot struct {
		Seed *int64 `json:"seed"`
		Tick *int   `json:"tick"`
	}

	var postSnapshot resetSnapshot
	if err := json.Unmarshal(postRec.Body.Bytes(), &postSnapshot); err != nil {
		t.Fatalf("не удалось прочитать снимок после POST: %v", err)
	}
	if postSnapshot.Seed == nil || *postSnapshot.Seed != 0 {
		t.Errorf("после POST ожидали seed=0. Ответ: %s", postRec.Body.String())
	}
	if postSnapshot.Tick == nil || *postSnapshot.Tick != 0 {
		t.Errorf("после POST ожидали tick=0. Ответ: %s", postRec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/state", nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("GET /api/state вернул %d, ожидали 200. Ответ: %s",
			getRec.Code, getRec.Body.String())
	}

	var getSnapshot resetSnapshot
	if err := json.Unmarshal(getRec.Body.Bytes(), &getSnapshot); err != nil {
		t.Fatalf("не удалось прочитать снимок после GET: %v", err)
	}
	if getSnapshot.Seed == nil || *getSnapshot.Seed != 0 {
		t.Errorf("после GET ожидали seed=0. Ответ: %s", getRec.Body.String())
	}
}
