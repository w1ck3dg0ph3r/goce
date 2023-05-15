package compilers

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Compiler interface {
	Info() (CompilerInfo, error)
	Compile(ctx context.Context, code []byte) (Result, error)
}

type CompilerInfo struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Platform string `json:"platform"`
}

type Result struct {
	CompilerInfo   CompilerInfo
	SourceFilename string
	SourceCode     []byte

	BuildOutput   []byte
	ObjdumpOutput []byte
}

// List returns available compilers.
func List() []CompilerInfo {
	infos := make([]CompilerInfo, 0, len(compilers))
	for _, comp := range compilers {
		info, err := comp.Info()
		if err != nil {
			continue
		}
		infos = append(infos, info)
	}
	return infos
}

// Get returns compiler with a given name.
func Get(name string) Compiler {
	if comp, ok := compilerByName[name]; ok {
		return comp
	}
	return nil
}

// Default returns default compiler.
func Default() Compiler {
	return defaultCompiler
}

var (
	compilers       []Compiler
	compilerByName  = map[string]Compiler{}
	defaultCompiler Compiler
)

func init() {
	registerGoFromPath()
	registerGoFromHomeSdk()
	if len(compilers) > 0 {
		defaultCompiler = compilers[0]
	}
}

func registerGoFromPath() {
	path, err := exec.LookPath("go")
	if err != nil {
		return
	}
	comp := &localCompiler{GoPath: path}
	info, err := comp.Info()
	if err != nil {
		return
	}
	compilers = append(compilers, comp)
	compilerByName[info.Name] = comp
}

func registerGoFromHomeSdk() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	goSdkDir := filepath.Join(homeDir, "sdk")
	entries, err := os.ReadDir(goSdkDir)
	if err != nil {
		return
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() > entries[j].Name() })
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "go") {
			continue
		}
		comp := &localCompiler{GoPath: filepath.Join(goSdkDir, e.Name(), "bin", "go")}
		info, err := comp.Info()
		if err != nil {
			continue
		}
		compilers = append(compilers, comp)
		compilerByName[info.Name] = comp
	}
}
