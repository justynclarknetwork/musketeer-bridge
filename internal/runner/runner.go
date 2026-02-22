package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"musketeer-bridge/internal/registry"
)

type RunRequest struct {
	Mode string                 `json:"mode"`
	Cwd  string                 `json:"cwd"`
	Args map[string]interface{} `json:"args"`
}

type RunResult struct {
	OK       bool        `json:"ok"`
	ExitCode int         `json:"exit_code"`
	Error    *ErrPayload `json:"error,omitempty"`
	Stdout   string      `json:"stdout,omitempty"`
	Stderr   string      `json:"stderr,omitempty"`
	StdoutJS any         `json:"stdout_json,omitempty"`
}

type ErrPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func codeErr(code, msg string, exit int) RunResult {
	return RunResult{OK: false, ExitCode: exit, Error: &ErrPayload{Code: code, Message: msg}}
}

func IsWithinRoots(cwd string, roots []string) bool {
	realCwd, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		return false
	}
	realCwd, _ = filepath.Abs(realCwd)
	for _, r := range roots {
		rr, err := filepath.EvalSymlinks(r)
		if err != nil {
			continue
		}
		rr, _ = filepath.Abs(rr)
		if strings.HasPrefix(realCwd, rr+string(os.PathSeparator)) || realCwd == rr {
			return true
		}
	}
	return false
}

func BuildArgv(spec registry.ToolSpec, req RunRequest) []string {
	argv := append([]string{}, spec.Exec.Argv...)
	for _, m := range spec.Exec.ArgsMap {
		v, ok := req.Args[m.Input]
		if !ok {
			continue
		}
		switch vv := v.(type) {
		case bool:
			if vv {
				if m.Kind == "flag" {
					argv = append(argv, m.Flag)
				} else {
					argv = append(argv, m.Flag, "true")
				}
			}
		case string:
			argv = append(argv, m.Flag, vv)
		case float64:
			argv = append(argv, m.Flag, strconv.FormatFloat(vv, 'f', -1, 64))
		case []interface{}:
			for _, x := range vv {
				argv = append(argv, m.Flag, fmt.Sprint(x))
			}
		default:
			argv = append(argv, m.Flag, fmt.Sprint(v))
		}
	}
	return argv
}

func ParseOneJSONObject(s string) (any, error) {
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if _, ok := v.(map[string]any); !ok {
		return nil, errors.New("not object")
	}
	var extra any
	if err := dec.Decode(&extra); err == nil {
		return nil, errors.New("multiple json values")
	}
	return v, nil
}

func Run(spec registry.ToolSpec, req RunRequest, roots []string, envAllow []string, timeoutMs int) RunResult {
	if !IsWithinRoots(req.Cwd, roots) {
		return codeErr("ERR_CWD_NOT_ALLOWLISTED", "cwd is not in allowlisted roots", 40)
	}
	argv := BuildArgv(spec, req)
	if len(argv) == 0 {
		return codeErr("ERR_EXEC_FAILED", "empty argv", 70)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = req.Cwd
	env := []string{}
	for _, k := range envAllow {
		if v, ok := os.LookupEnv(k); ok {
			env = append(env, k+"="+v)
		}
	}
	cmd.Env = env
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	out := outb.String()
	errOut := errb.String()
	if ctx.Err() == context.DeadlineExceeded {
		return codeErr("ERR_TIMEOUT", "command timed out", 124)
	}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return RunResult{OK: false, ExitCode: ee.ExitCode(), Error: &ErrPayload{Code: "ERR_EXEC_FAILED", Message: "command failed"}, Stdout: out, Stderr: errOut}
		}
		return codeErr("ERR_EXEC_FAILED", "tool execution failed", 70)
	}
	res := RunResult{OK: true, ExitCode: 0, Stdout: out, Stderr: errOut}
	if spec.JsonMode && req.Mode == "json" {
		obj, jerr := ParseOneJSONObject(out)
		if jerr != nil {
			return codeErr("ERR_STDOUT_NOT_JSON", "stdout is not exactly one JSON object", 40)
		}
		res.StdoutJS = obj
	}
	return res
}
