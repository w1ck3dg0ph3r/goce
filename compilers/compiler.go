package compilers

import (
	"io"
	"os/exec"
)

type Compiler interface {
	Info() (CompilerInfo, error)
	Compile(code []byte) (Result, error)
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

	BuildOutput   io.Reader
	ObjdumpOutput io.Reader
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
