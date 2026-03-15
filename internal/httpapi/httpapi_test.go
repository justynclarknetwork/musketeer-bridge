package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"musketeer-bridge/internal/config"
	"musketeer-bridge/internal/httpapi"
	"musketeer-bridge/internal/logstore"
	"musketeer-bridge/internal/registry"
)

func makeAPI(t *testing.T) *httpapi.API {
	t.Helper()
	return &httpapi.API{
		Cfg: config.Default(),
		Reg: registry.Registry{Tools: map[string]registry.ToolSpec{}},
		Log: logstore.LogWriter{RunsDir: t.TempDir()},
	}
}

func TestHealthCheck(t *testing.T) {
	api := makeAPI(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["ok"] != true {
		t.Fatalf("expected ok=true, got %v", body["ok"])
	}
	if _, ok := body["exit_code"]; !ok {
		t.Fatal("missing exit_code in health response")
	}
	if body["exit_code"] != float64(0) {
		t.Fatalf("expected exit_code=0, got %v", body["exit_code"])
	}
}

func TestInvalidJSONBodyReturns400(t *testing.T) {
	api := makeAPI(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/tools/fake/run", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	errObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object, got %v", body)
	}
	if errObj["code"] != "ERR_INVALID_INPUT" {
		t.Fatalf("expected ERR_INVALID_INPUT, got %v", errObj["code"])
	}
	if body["exit_code"] != float64(40) {
		t.Fatalf("expected exit_code=40, got %v", body["exit_code"])
	}
}

func TestToolsListEmpty(t *testing.T) {
	api := makeAPI(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/tools", nil)
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["exit_code"]; !ok {
		t.Fatal("missing exit_code in tools list response")
	}
	tools, ok := body["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools array, got %v", body["tools"])
	}
	if len(tools) != 0 {
		t.Fatalf("expected empty tools list, got %v", tools)
	}
}

func TestToolNotFound(t *testing.T) {
	api := makeAPI(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/tools/does-not-exist", nil)
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	errObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object in response, got %v", body)
	}
	if errObj["code"] != "ERR_TOOL_NOT_FOUND" {
		t.Fatalf("expected ERR_TOOL_NOT_FOUND, got %v", errObj["code"])
	}
}

func TestRunToolNotFoundReturns404(t *testing.T) {
	api := makeAPI(t)
	body := `{"version":"0.1.0","args":{},"cwd":"/tmp","env":{},"mode":"json","client":{"name":"test"}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/tools/does-not-exist/run", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	errObj, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object, got %v", resp)
	}
	if errObj["code"] != "ERR_TOOL_NOT_FOUND" {
		t.Fatalf("expected ERR_TOOL_NOT_FOUND, got %v", errObj["code"])
	}
}

func TestResponseAlwaysIncludesExitCode(t *testing.T) {
	api := makeAPI(t)
	paths := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/v1/health"},
		{http.MethodGet, "/v1/tools"},
		{http.MethodGet, "/v1/tools/does-not-exist"},
	}
	for _, p := range paths {
		req := httptest.NewRequest(p.method, p.path, nil)
		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		var body map[string]any
		if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
			t.Fatalf("%s %s: %v", p.method, p.path, err)
		}
		if _, ok := body["exit_code"]; !ok {
			t.Fatalf("%s %s: missing exit_code", p.method, p.path)
		}
	}
}
