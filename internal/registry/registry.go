package registry

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
)

type ArgMap struct {
	Input    string `json:"input"`
	Flag     string `json:"flag"`
	Kind     string `json:"kind"`
	Repeated bool   `json:"repeated"`
}

type ExecSpec struct {
	Argv       []string `json:"argv"`
	ArgsMap    []ArgMap `json:"args_mapping"`
	WorkingDir string   `json:"working_dir"`
}

type ToolSpec struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	JsonMode    bool     `json:"json_mode"`
	Exec        ExecSpec `json:"exec"`
}

type Registry struct {
	Tools map[string]ToolSpec
}

func Load(base string) (Registry, error) {
	reg := Registry{Tools: map[string]ToolSpec{}}
	toolsDir := filepath.Join(base, "tools")
	entries, err := os.ReadDir(toolsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return reg, nil
		}
		return reg, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		vdir := filepath.Join(toolsDir, name)
		vers, _ := os.ReadDir(vdir)
		names := []string{}
		for _, v := range vers {
			if v.IsDir() {
				names = append(names, v.Name())
			}
		}
		sort.Strings(names)
		if len(names) == 0 {
			continue
		}
		latest := names[len(names)-1]
		p := filepath.Join(vdir, latest, "tool.json")
		b, err := os.ReadFile(p)
		if err != nil {
			return reg, errors.New("ERR_REGISTRY_INVALID")
		}
		var t ToolSpec
		if err := json.Unmarshal(b, &t); err != nil {
			return reg, errors.New("ERR_REGISTRY_INVALID")
		}
		if t.Name == "" || t.Version == "" || t.Description == "" || len(t.Exec.Argv) == 0 {
			return reg, errors.New("ERR_REGISTRY_INVALID")
		}
		reg.Tools[name] = t
	}
	return reg, nil
}
