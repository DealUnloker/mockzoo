package integration_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

type petResponse struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Status    string   `json:"status"`
	Species   string   `json:"species"`
	Breed     string   `json:"breed"`
	PhotoURL  string   `json:"photoUrl"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"createdAt"`
}

type petListResponse struct {
	Items []petResponse `json:"items"`
	Total int           `json:"total"`
}

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Seed pets (ids 1-3) are present via listPets.
func TestPetListSeeds(t *testing.T) {
	resp := doGet(t, basePathV1+"/pets?limit=100")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	listed := parseJSON[petListResponse](t, resp)

	if listed.Total < 3 {
		t.Fatalf("expected total >= 3, got %d", listed.Total)
	}

	seen := map[int64]bool{}
	for _, p := range listed.Items {
		seen[p.ID] = true
	}

	for _, id := range []int64{1, 2, 3} {
		if !seen[id] {
			t.Errorf("expected seed pet id %d in list, not found", id)
		}
	}
}

// getPetById(1) returns Barsik with required fields.
func TestGetPetByIDSeed(t *testing.T) {
	resp := doGet(t, basePathV1+"/pets/1")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	p := parseJSON[petResponse](t, resp)

	if p.Name != "Barsik" {
		t.Errorf("expected name Barsik, got %q", p.Name)
	}

	if p.Species != "cat" {
		t.Errorf("expected species cat, got %q", p.Species)
	}

	if p.Status == "" || p.CreatedAt == "" {
		t.Errorf("expected status and createdAt to be set, got status=%q createdAt=%q", p.Status, p.CreatedAt)
	}
}

// 404 for a non-existent pet id, with error.code pet_not_found.
func TestGetPetByIDNotFound(t *testing.T) {
	resp := doGet(t, basePathV1+"/pets/99999")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	errBody := parseJSON[errorResponse](t, resp)

	if errBody.Error.Code != "pet_not_found" {
		t.Errorf("expected error.code pet_not_found, got %q", errBody.Error.Code)
	}
}

// Create a pet, patch its status, then delete it; ids start at 1001.
func TestPetCreateUpdateDelete(t *testing.T) {
	createBody := `{"name":"Integration Pup","species":"dog"}`

	createResp := doPost(t, basePathV1+"/pets", []byte(createBody))
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", createResp.StatusCode)
	}

	created := parseJSON[petResponse](t, createResp)

	if created.ID < 1001 {
		t.Errorf("expected created id >= 1001, got %d", created.ID)
	}

	patchBody := `{"status":"sold"}`

	patchResp := doPatch(t, basePathV1+"/pets/"+itoa(created.ID), []byte(patchBody))
	defer patchResp.Body.Close()

	if patchResp.StatusCode != http.StatusOK {
		t.Fatalf("patch: expected 200, got %d", patchResp.StatusCode)
	}

	patched := parseJSON[petResponse](t, patchResp)

	if patched.Status != "sold" {
		t.Errorf("expected status sold, got %q", patched.Status)
	}

	delResp := doDelete(t, basePathV1+"/pets/"+itoa(created.ID))
	defer delResp.Body.Close()

	if delResp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", delResp.StatusCode)
	}

	getResp := doGet(t, basePathV1+"/pets/"+itoa(created.ID))
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusNotFound {
		t.Fatalf("get after delete: expected 404, got %d", getResp.StatusCode)
	}
}

// resetSandbox restores Barsik and removes pets created since.
func TestResetSandbox(t *testing.T) {
	createBody := `{"name":"Temp Pet","species":"fish"}`

	createResp := doPost(t, basePathV1+"/pets", []byte(createBody))
	defer createResp.Body.Close()

	created := parseJSON[petResponse](t, createResp)

	resetResp := doPost(t, basePathV1+"/admin/reset", nil)
	defer resetResp.Body.Close()

	if resetResp.StatusCode != http.StatusOK {
		t.Fatalf("reset: expected 200, got %d", resetResp.StatusCode)
	}

	var resetBody struct {
		Reset bool `json:"reset"`
		Pets  int  `json:"pets"`
	}

	if err := json.NewDecoder(resetResp.Body).Decode(&resetBody); err != nil {
		t.Fatalf("reset: failed to decode response: %v", err)
	}

	if !resetBody.Reset || resetBody.Pets != 3 {
		t.Errorf("expected {reset:true, pets:3}, got %+v", resetBody)
	}

	barsikResp := doGet(t, basePathV1+"/pets/1")
	defer barsikResp.Body.Close()

	barsik := parseJSON[petResponse](t, barsikResp)

	if barsik.Name != "Barsik" {
		t.Errorf("expected Barsik restored, got %q", barsik.Name)
	}

	goneResp := doGet(t, basePathV1+"/pets/"+itoa(created.ID))
	defer goneResp.Body.Close()

	if goneResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected pet created before reset to be gone, got status %d", goneResp.StatusCode)
	}
}

// GET /v1/openapi.json returns valid JSON with an "openapi" key.
func TestOpenAPIJSON(t *testing.T) {
	resp := doGet(t, basePathV1+"/openapi.json")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var doc map[string]any

	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		t.Fatalf("failed to decode openapi.json: %v", err)
	}

	if _, ok := doc["openapi"]; !ok {
		t.Error("expected \"openapi\" key in the spec")
	}
}

// CORS header is present on a cross-origin GET response. The middleware only
// answers requests that carry an Origin header, so the test must send one.
func TestCORSHeaderPresent(t *testing.T) {
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, basePathV1+"/pets", http.NoBody)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	req.Header.Set("Origin", "http://localhost:3000")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", basePathV1+"/pets", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("Access-Control-Allow-Origin") == "" {
		t.Error("expected Access-Control-Allow-Origin header to be set")
	}
}

func itoa(id int64) string {
	return strconv.FormatInt(id, 10)
}
