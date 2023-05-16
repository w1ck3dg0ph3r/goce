package compilers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Compiler interface {
	Info() (CompilerInfo, error)
	Compile(ctx context.Context, config Config, code []byte) (Result, error)
}

type CompilerInfo struct {
	Version      string `json:"version"`
	Platform     string `json:"platform"`
	Architecture string `json:"architecture"`
}

type CompilerVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

type Config struct {
	Platform     string `json:"platform"`
	Architecture string `json:"architecture"`
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
	for _, desc := range compilerDescs {
		infos = append(infos, desc.Info)
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

func ParseInfo(name string) (CompilerInfo, error) {
	var ci CompilerInfo
	match := reCompilerName.FindStringSubmatch(name)
	if match == nil {
		return ci, fmt.Errorf("invalid compiler name: %s", name)
	}
	ci = CompilerInfo{
		Version:      match[reCompilerName_Version],
		Platform:     match[reCompilerName_Platform],
		Architecture: match[reCompilerName_Architecture],
	}
	return ci, nil
}

func (i CompilerInfo) Name() string {
	return fmt.Sprintf("go%s %s/%s", i.Version, i.Platform, i.Architecture)
}

var (
	compilers       []Compiler
	compilerDescs   []compilerDesc
	compilerByName  = map[string]Compiler{}
	defaultCompiler Compiler
)

type compilerDesc struct {
	Name     string
	Info     CompilerInfo
	Compiler Compiler
}

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
	registerLocalCompiler(path)
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
		registerLocalCompiler(filepath.Join(goSdkDir, e.Name(), "bin", "go"))
	}
}

func registerLocalCompiler(path string) {
	comp := &localCompiler{GoPath: path}
	info, err := comp.Info()
	if err != nil {
		return
	}
	compilers = append(compilers, comp)
	desc := compilerDesc{
		Name:     info.Name(),
		Info:     info,
		Compiler: comp,
	}
	compilerDescs = append(compilerDescs, desc)
	compilerByName[desc.Name] = comp
	addArchitectures(comp, desc)
}

func addArchitectures(comp Compiler, desc compilerDesc) {
	var supportedArchs []string
	if desc.Info.Platform == "linux" {
		supportedArchs = append(supportedArchs, "amd64", "386", "arm64", "arm", "ppc64")
	}
	for _, arch := range supportedArchs {
		if arch == desc.Info.Architecture {
			continue
		}
		desc.Info.Architecture = arch
		desc.Name = desc.Info.Name()
		compilerDescs = append(compilerDescs, desc)
		compilerByName[desc.Name] = comp
	}
}

var reCompilerName = regexp.MustCompile(`go(\d+\.\d+(\.\d+)?)\s+(\w+)/(\w+)`)

const (
	reCompilerName_Version = iota + 1
	reCompilerName_VersionPatch
	reCompilerName_Platform
	reCompilerName_Architecture
)
