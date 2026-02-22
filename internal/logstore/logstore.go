package logstore

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type LogWriter struct{ RunsDir string }

func (l LogWriter) NewRunDir() (string, string, error) {
	now := time.Now().UTC()
	runID := fmt.Sprintf("%s-%06d", now.Format("20060102T150405.000Z"), rand.Intn(1000000))
	dir := filepath.Join(l.RunsDir, now.Format("2006"), now.Format("01"), now.Format("02"), runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	return runID, dir, nil
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func (l LogWriter) WriteAll(dir string, req any, resolved any, stdoutJSON any, stderr string, result any) {
	_ = writeJSON(filepath.Join(dir, "request.json"), req)
	_ = writeJSON(filepath.Join(dir, "resolved.json"), resolved)
	if stdoutJSON != nil {
		_ = writeJSON(filepath.Join(dir, "stdout.json"), stdoutJSON)
	}
	_ = os.WriteFile(filepath.Join(dir, "stderr.txt"), []byte(stderr), 0o644)
	_ = writeJSON(filepath.Join(dir, "result.json"), result)
}
