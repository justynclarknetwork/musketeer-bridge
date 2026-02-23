package contracttest

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"musketeer-bridge/internal/config"
	"musketeer-bridge/internal/httpapi"
	"musketeer-bridge/internal/logstore"
	"musketeer-bridge/internal/registry"
)

type runResp struct {
	ExitCode int `json:"exit_code"`
	Error    *struct {
		Code string `json:"code"`
	} `json:"error,omitempty"`
	StdoutJSON map[string]interface{} `json:"stdout_json,omitempty"`
	Stderr     string                 `json:"stderr,omitempty"`
}

func buildFakeCLI(t *testing.T, outPath string) {
	t.Helper()
	cmd := exec.Command("go", "build", "-o", outPath, "./internal/contracttest/fakecli")
	cmd.Dir = filepath.Clean(filepath.Join("..", ".."))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build fakecli failed: %v\n%s", err, string(out))
	}
}

func buildBridgeCLI(t *testing.T, outPath string) {
	t.Helper()
	cmd := exec.Command("go", "build", "-o", outPath, "./cmd/musketeer-bridge")
	cmd.Dir = filepath.Clean(filepath.Join("..", ".."))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build bridge failed: %v\n%s", err, string(out))
	}
}

func writeToolSpec(t *testing.T, registryDir, fakeCLI, behavior string) {
	t.Helper()
	dir := filepath.Join(registryDir, "tools", "fake", "0.1.0")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	spec := map[string]interface{}{
		"name":        "fake",
		"version":     "0.1.0",
		"description": "fake test tool",
		"json_mode":   true,
		"exec": map[string]interface{}{
			"argv":         []string{fakeCLI, behavior},
			"args_mapping": []interface{}{},
		},
	}
	b, _ := json.MarshalIndent(spec, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "tool.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func startServer(t *testing.T, workdir, behavior string, maxRuntime int) (*httptest.Server, string) {
	t.Helper()
	home := t.TempDir()
	registryDir := filepath.Join(home, ".musketeer", "registry")
	runsDir := filepath.Join(home, ".musketeer", "runs")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatal(err)
	}
	fakeCLIPath := filepath.Join(home, "fakecli")
	buildFakeCLI(t, fakeCLIPath)
	writeToolSpec(t, registryDir, fakeCLIPath, behavior)
	reg, err := registry.Load(registryDir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		ListenAddr:       "127.0.0.1:0",
		AllowlistedRoots: []string{workdir},
		EnvAllowlist:     []string{"PATH", "HOME", "USER", "SHELL", "TERM"},
		MaxRuntimeMs:     maxRuntime,
		RegistryDir:      registryDir,
		RunsDir:          runsDir,
	}
	api := &httpapi.API{Cfg: cfg, Reg: reg, Log: logstore.LogWriter{RunsDir: runsDir}}
	return httptest.NewServer(api), runsDir
}

func postRun(t *testing.T, url, cwd string) runResp {
	t.Helper()
	body := []byte(`{"version":"0.1.0","args":{},"cwd":"` + cwd + `","env":{},"mode":"json","client":{"name":"test"}}`)
	resp, err := http.Post(url+"/v1/tools/fake/run", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var rr runResp
	if err := json.NewDecoder(resp.Body).Decode(&rr); err != nil {
		t.Fatal(err)
	}
	return rr
}

func latestRunDir(t *testing.T, runsDir string) string {
	t.Helper()
	var latest string
	_ = filepath.WalkDir(runsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			latest = path
		}
		return nil
	})
	if latest == "" {
		t.Fatal("no run dir found")
	}
	return latest
}

func TestContractGoodJSON(t *testing.T) {
	workdir := t.TempDir()
	srv, runsDir := startServer(t, workdir, "good-json", 1000)
	defer srv.Close()
	r := postRun(t, srv.URL, workdir)
	if r.ExitCode != 0 || r.StdoutJSON == nil || r.StdoutJSON["ok"] != true || r.Stderr != "" {
		t.Fatalf("unexpected response: %+v", r)
	}
	rd := latestRunDir(t, runsDir)
	for _, f := range []string{"request.json", "resolved.json", "result.json", "stdout.json"} {
		if _, err := os.Stat(filepath.Join(rd, f)); err != nil {
			t.Fatal(err)
		}
	}
	var obj map[string]interface{}
	b, _ := os.ReadFile(filepath.Join(rd, "stdout.json"))
	if err := json.Unmarshal(b, &obj); err != nil {
		t.Fatal(err)
	}
}

func TestContractBadJSONText(t *testing.T) {
	workdir := t.TempDir()
	srv, runsDir := startServer(t, workdir, "bad-json-text", 1000)
	defer srv.Close()
	r := postRun(t, srv.URL, workdir)
	if r.Error == nil || r.Error.Code != "ERR_STDOUT_NOT_JSON" {
		t.Fatalf("expected ERR_STDOUT_NOT_JSON, got %+v", r)
	}
	rd := latestRunDir(t, runsDir)
	if _, err := os.Stat(filepath.Join(rd, "result.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(rd, "stdout.json")); err == nil {
		t.Fatal("stdout.json should not exist")
	}
}

func TestContractBadJSONMulti(t *testing.T) {
	workdir := t.TempDir()
	srv, runsDir := startServer(t, workdir, "bad-json-multi", 1000)
	defer srv.Close()
	r := postRun(t, srv.URL, workdir)
	if r.Error == nil || r.Error.Code != "ERR_STDOUT_NOT_JSON" {
		t.Fatalf("expected ERR_STDOUT_NOT_JSON, got %+v", r)
	}
	rd := latestRunDir(t, runsDir)
	if _, err := os.Stat(filepath.Join(rd, "result.json")); err != nil {
		t.Fatal(err)
	}
}

func TestContractBadJSONArray(t *testing.T) {
	workdir := t.TempDir()
	srv, runsDir := startServer(t, workdir, "bad-json-array", 1000)
	defer srv.Close()
	r := postRun(t, srv.URL, workdir)
	if r.Error == nil || r.Error.Code != "ERR_STDOUT_NOT_JSON" {
		t.Fatalf("expected ERR_STDOUT_NOT_JSON, got %+v", r)
	}
	rd := latestRunDir(t, runsDir)
	if _, err := os.Stat(filepath.Join(rd, "result.json")); err != nil {
		t.Fatal(err)
	}
}

func TestContractTimeout(t *testing.T) {
	workdir := t.TempDir()
	srv, runsDir := startServer(t, workdir, "hang", 100)
	defer srv.Close()
	r := postRun(t, srv.URL, workdir)
	if r.Error == nil || r.Error.Code != "ERR_TIMEOUT" || r.ExitCode == 0 {
		t.Fatalf("expected timeout with nonzero exit, got %+v", r)
	}
	rd := latestRunDir(t, runsDir)
	if _, err := os.Stat(filepath.Join(rd, "result.json")); err != nil {
		t.Fatal(err)
	}
}

func TestContractAllowlistRejected(t *testing.T) {
	workdir := t.TempDir()
	srv, runsDir := startServer(t, workdir, "good-json", 1000)
	defer srv.Close()
	outside := t.TempDir()
	r := postRun(t, srv.URL, outside)
	if r.Error == nil || r.Error.Code != "ERR_CWD_NOT_ALLOWLISTED" || r.ExitCode == 0 {
		t.Fatalf("expected ERR_CWD_NOT_ALLOWLISTED with nonzero exit, got %+v", r)
	}
	rd := latestRunDir(t, runsDir)
	if _, err := os.Stat(filepath.Join(rd, "result.json")); err != nil {
		t.Fatal(err)
	}
}

func TestContractHelpDoesNotBind(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "musketeer-bridge")
	buildBridgeCLI(t, bin)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	cmd := exec.Command(bin, "--help")
	cmd.Env = append(os.Environ(), "MUSKETEER_BRIDGE_LISTEN_ADDR="+addr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help failed: %v\n%s", err, string(out))
	}
	text := string(out)
	if !strings.Contains(text, "Usage:") {
		t.Fatalf("missing usage marker: %s", text)
	}
	if strings.Contains(strings.ToLower(text), "listening on") {
		t.Fatalf("help emitted listening log: %s", text)
	}

	ln2, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("expected port free after help, bind failed: %v", err)
	}
	_ = ln2.Close()
}
