package compilers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type Config struct {
	SearchGoPath  bool // Search PATH for go compilers
	SearchSDKPath bool // Search ~/sdk/go* for go compilers

	LocalCompilers []string // Paths of local go compiler executables

	AdditionalArchitectures bool // Add supported cross-compilation architectures
}

var ErrNoCompilers = errors.New("no compilers found")

// New creates and initializes [CompilersSvc].
func New(cfg *Config) (*CompilersSvc, error) {
	svc := &CompilersSvc{
		cfg: cfg,
	}
	if err := svc.registerCompilers(); err != nil {
		return nil, err
	}
	return svc, nil
}

type CompilersSvc struct {
	cfg *Config

	compilers       []*compilerDesc
	compilerByName  map[string]*compilerDesc
	defaultCompiler *compilerDesc
}

type Compiler interface {
	Info() (CompilerInfo, error)
	Compile(ctx context.Context, cfg CompilerConfig, code []byte) (Result, error)
}

type CompilerVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

type CompilerConfig struct {
	Platform     string          `json:"platform"`
	Architecture string          `json:"architecture"`
	Options      CompilerOptions `json:"options"`
}

type CompilerOptions struct {
	DisableInlining      bool   `json:"disableInlining"`
	DisableOptimizations bool   `json:"disableOptimizations"`
	ArchitectureLevel    string `json:"architectureLevel"`
}

type Result struct {
	CompilerInfo   CompilerInfo `json:"compilerInfo"`
	SourceFilename string       `json:"sourceFilename"`
	SourceCode     []byte       `json:"sourceCode"`

	BuildOutput   []byte `json:"buildOutput"`
	ObjdumpOutput []byte `json:"objdumpOutput"`
}

// List returns available compilers.
func (svc *CompilersSvc) List() []CompilerInfo {
	infos := make([]CompilerInfo, 0, len(svc.compilers))
	for _, desc := range svc.compilers {
		infos = append(infos, desc.Info)
	}
	return infos
}

// Get returns compiler with a given name.
func (svc *CompilersSvc) Get(name string) Compiler {
	if comp, ok := svc.compilerByName[name]; ok {
		return comp.Compiler
	}
	return nil
}

// Default returns default compiler.
func (svc *CompilersSvc) Default() Compiler {
	return svc.defaultCompiler.Compiler
}

type CompilerInfo struct {
	Version      string `json:"version"`
	Platform     string `json:"platform"`
	Architecture string `json:"architecture"`
}

// ParseInfo parses [CompilerInfo] from compiler name.
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

type compilerDesc struct {
	Name     string
	Info     CompilerInfo
	Compiler Compiler

	version *semver.Version
}

func (svc *CompilersSvc) registerCompilers() error {
	svc.compilers = nil
	svc.compilerByName = make(map[string]*compilerDesc)

	if svc.cfg.SearchGoPath {
		svc.registerGoFromPath()
	}
	if svc.cfg.SearchSDKPath {
		svc.registerGoFromHomeSdk()
	}
	for _, path := range svc.cfg.LocalCompilers {
		if err := svc.registerLocalCompiler(path); err != nil {
			return err
		}
	}

	sort.Slice(svc.compilers, func(i, j int) bool {
		if svc.compilers[i].version.Equal(svc.compilers[j].version) {
			io := architectureOrder[svc.compilers[i].Info.Architecture]
			jo := architectureOrder[svc.compilers[j].Info.Architecture]
			return io < jo
		}
		return svc.compilers[i].version.GreaterThan(svc.compilers[j].version)
	})

	if len(svc.compilers) == 0 {
		return ErrNoCompilers
	}
	svc.defaultCompiler = svc.compilers[0]
	return nil
}

func (svc *CompilersSvc) registerGoFromPath() {
	path, err := exec.LookPath("go")
	if err != nil {
		return
	}
	if err := svc.registerLocalCompiler(path); err != nil {
		return
	}
}

func (svc *CompilersSvc) registerGoFromHomeSdk() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	goSdkDir := filepath.Join(homeDir, "sdk")
	entries, err := os.ReadDir(goSdkDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "go") {
			continue
		}
		if err := svc.registerLocalCompiler(filepath.Join(goSdkDir, e.Name(), "bin", "go")); err != nil {
			return
		}
	}
}

func (svc *CompilersSvc) registerLocalCompiler(path string) error {
	fs, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("register compiler: %w", err)
	}
	if fs.Mode().Perm()&0o111 == 0 {
		return fmt.Errorf("register compiler: not executable: %s", path)
	}

	comp := &localCompiler{GoPath: path}
	info, err := comp.Info()
	if err != nil {
		return err
	}
	desc := &compilerDesc{
		Name:     info.Name(),
		Info:     info,
		Compiler: comp,
	}
	desc.version, err = semver.NewVersion(desc.Info.Version)
	if err != nil {
		return fmt.Errorf("register compiler: invalid version: %w", err)
	}
	if _, exists := svc.compilerByName[desc.Name]; exists {
		return nil
	}
	svc.compilers = append(svc.compilers, desc)
	svc.compilerByName[desc.Name] = desc
	if svc.cfg.AdditionalArchitectures {
		svc.addArchitectures(comp, desc)
	}
	return nil
}

func (svc *CompilersSvc) addArchitectures(comp Compiler, desc *compilerDesc) {
	var supportedArchs []string
	if desc.Info.Platform == "linux" {
		supportedArchs = append(supportedArchs, "amd64", "386", "arm64", "arm", "ppc64")
	}
	for _, arch := range supportedArchs {
		if arch == desc.Info.Architecture {
			continue
		}
		newDesc := *desc
		newDesc.Info.Architecture = arch
		newDesc.Name = newDesc.Info.Name()
		svc.compilers = append(svc.compilers, &newDesc)
		svc.compilerByName[newDesc.Name] = &newDesc
	}
}

var reCompilerName = regexp.MustCompile(`go(\d+\.\d+(\.\d+)?)\s+(\w+)/(\w+)`)

const (
	reCompilerName_Version = iota + 1
	reCompilerName_VersionPatch
	reCompilerName_Platform
	reCompilerName_Architecture
)

var architectureOrder = map[string]int{
	"amd64": 1,
	"arm64": 2,
	"ppc64": 3,
	"386":   4,
	"arm":   5,
}
