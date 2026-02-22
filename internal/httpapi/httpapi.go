package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"musketeer-bridge/internal/config"
	"musketeer-bridge/internal/logstore"
	"musketeer-bridge/internal/registry"
	"musketeer-bridge/internal/runner"
)

type API struct {
	Cfg config.Config
	Reg registry.Registry
	Log logstore.LogWriter
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/v1/health" {
		writeJSON(w, 200, map[string]any{"ok": true, "exit_code": 0})
		return
	}
	if r.Method == http.MethodGet && r.URL.Path == "/v1/tools" {
		tools := []string{}
		for n := range a.Reg.Tools {
			tools = append(tools, n)
		}
		writeJSON(w, 200, map[string]any{"tools": tools, "exit_code": 0})
		return
	}
	if strings.HasPrefix(r.URL.Path, "/v1/tools/") {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/tools/"), "/")
		name := parts[0]
		spec, ok := a.Reg.Tools[name]
		if !ok {
			res := map[string]any{"exit_code": 40, "error": map[string]any{"code": "ERR_TOOL_NOT_FOUND", "message": "tool not found"}}
			writeJSON(w, 404, res)
			return
		}
		if len(parts) == 1 && r.Method == http.MethodGet {
			writeJSON(w, 200, map[string]any{"tool": spec, "exit_code": 0})
			return
		}
		if len(parts) == 2 && parts[1] == "run" && r.Method == http.MethodPost {
			var req runner.RunRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				res := map[string]any{"exit_code": 40, "error": map[string]any{"code": "ERR_INVALID_INPUT", "message": "invalid json"}}
				writeJSON(w, 400, res)
				return
			}
			runID, dir, _ := a.Log.NewRunDir()
			result := runner.Run(spec, req, a.Cfg.AllowlistedRoots, a.Cfg.EnvAllowlist, a.Cfg.MaxRuntimeMs)
			resp := map[string]any{"run_id": runID, "exit_code": result.ExitCode, "ok": result.OK, "stdout": result.Stdout, "stderr": result.Stderr}
			if result.StdoutJS != nil {
				resp["stdout_json"] = result.StdoutJS
			}
			if result.Error != nil {
				resp["error"] = result.Error
			}
			a.Log.WriteAll(dir, req, spec, result.StdoutJS, result.Stderr, resp)
			status := 200
			if result.Error != nil {
				status = 400
				if result.Error.Code == "ERR_TOOL_NOT_FOUND" {
					status = 404
				}
				if result.ExitCode == 70 {
					status = 500
				}
			}
			writeJSON(w, status, resp)
			return
		}
	}
	writeJSON(w, 404, map[string]any{"exit_code": 40, "error": map[string]any{"code": "ERR_NOT_FOUND", "message": "not found"}})
}
