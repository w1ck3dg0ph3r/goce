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
	"sync"
	"time"

	"github.com/Masterminds/semver/v3"
)

type Config struct {
	SearchGoPath   bool     // Search PATH for go compilers.
	SearchSDKPath  bool     // Search ~/sdk/go* for go compilers.
	LocalCompilers []string // Paths of local go compiler executables.

	AdditionalArchitectures bool // Add supported cross-compilation architectures.

	EnableModules bool // Enable modules support.
}

var (
	ErrNoCompilers = errors.New("no compilers found")
	ErrInvalidName = errors.New("invalid compiler name")
	ErrInvalidPath = errors.New("invalid compiler path")
	ErrBuildFailed = errors.New("build failed")
)

// New creates and initializes [Service].
func New(cfg *Config) (*Service, error) {
	svc := &Service{
		cfg: cfg,
	}
	if err := svc.refreshAvailable(); err != nil {
		return nil, err
	}
	return svc, nil
}

type Service struct {
	cfg *Config

	availableMu  sync.RWMutex
	available    availableCompilers
	availableTTL time.Time
}

type availableCompilers struct {
	cfg             *Config
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

	BuildOutput []byte `json:"buildOutput"`
}

// List returns available compilers.
func (svc *Service) List() []CompilerInfo {
	_ = svc.refreshAvailable()
	svc.availableMu.RLock()
	defer svc.availableMu.RUnlock()
	infos := make([]CompilerInfo, 0, len(svc.available.compilers))
	for _, desc := range svc.available.compilers {
		infos = append(infos, desc.Info)
	}
	return infos
}

// Get returns compiler with a given name.
func (svc *Service) Get(name string) Compiler {
	_ = svc.refreshAvailable()
	svc.availableMu.RLock()
	defer svc.availableMu.RUnlock()
	if comp, ok := svc.available.compilerByName[name]; ok {
		return comp.Compiler
	}
	return nil
}

// Default returns default compiler.
func (svc *Service) Default() Compiler {
	_ = svc.refreshAvailable()
	svc.availableMu.RLock()
	defer svc.availableMu.RUnlock()
	if svc.available.defaultCompiler == nil {
		return nil
	}
	return svc.available.defaultCompiler.Compiler
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
		return ci, fmt.Errorf("%w: %s", ErrInvalidName, name)
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

func (svc *Service) refreshAvailable() error {
	now := time.Now()
	needRefresh := false
	svc.availableMu.RLock()
	if svc.availableTTL.Before(now) {
		needRefresh = true
	}
	svc.availableMu.RUnlock()

	if !needRefresh {
		return nil
	}

	svc.availableMu.Lock()
	defer svc.availableMu.Unlock()

	var err error
	svc.available, err = svc.listAvailable()
	if err != nil {
		return err
	}
	svc.availableTTL = time.Now().Add(15 * time.Second)
	return nil
}

func (svc *Service) listAvailable() (availableCompilers, error) {
	ac := availableCompilers{
		cfg:            svc.cfg,
		compilerByName: map[string]*compilerDesc{},
	}

	if svc.cfg.SearchGoPath {
		ac.searchGoPath()
	}
	if svc.cfg.SearchSDKPath {
		ac.searchSDKPath()
	}
	for _, path := range svc.cfg.LocalCompilers {
		if err := ac.addLocal(path); err != nil {
			return availableCompilers{}, err
		}
	}

	if svc.cfg.AdditionalArchitectures {
		for _, cd := range ac.compilers {
			ac.addArchitectures(cd)
		}
	}

	sort.Slice(ac.compilers, func(i, j int) bool {
		if ac.compilers[i].version.Equal(ac.compilers[j].version) {
			io := architectureOrder[ac.compilers[i].Info.Architecture]
			jo := architectureOrder[ac.compilers[j].Info.Architecture]
			return io < jo
		}
		return ac.compilers[i].version.GreaterThan(ac.compilers[j].version)
	})

	if len(ac.compilers) > 0 {
		ac.defaultCompiler = ac.compilers[0]
	}

	return ac, nil
}

func (ac *availableCompilers) searchGoPath() {
	path, err := exec.LookPath("go")
	if err != nil {
		return
	}
	if err := ac.addLocal(path); err != nil {
		return
	}
}

func (ac *availableCompilers) searchSDKPath() {
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
		if err := ac.addLocal(filepath.Join(goSdkDir, e.Name(), "bin", "go")); err != nil {
			return
		}
	}
}

func (ac *availableCompilers) addLocal(path string) error {
	fs, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidPath, err)
	}
	if fs.Mode().Perm()&0o111 == 0 {
		return fmt.Errorf("%w: not executable: %s", ErrInvalidPath, path)
	}

	comp := &localCompiler{
		GoPath:        path,
		EnableModules: ac.cfg.EnableModules,
	}
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
	if _, exists := ac.compilerByName[desc.Name]; exists {
		return nil
	}
	ac.compilers = append(ac.compilers, desc)
	ac.compilerByName[desc.Name] = desc
	return nil
}

func (ac *availableCompilers) addArchitectures(desc *compilerDesc) {
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
		ac.compilers = append(ac.compilers, &newDesc)
		ac.compilerByName[newDesc.Name] = &newDesc
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
