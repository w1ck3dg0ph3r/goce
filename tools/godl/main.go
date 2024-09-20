package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Flags struct {
	Help        bool `short:"h" long:"help" usage:"Display help message"`
	NumVersions int  `short:"n" long:"num-versions" default:"3" usage:"Number of latest go verions"`
}

func main() {
	var flags Flags
	ParseFlags(&flags)
	if flags.Help {
		flag.Usage()
		os.Exit(0)
	}

	fmt.Printf("listing installed verions\n")
	local, err := GetLocalVersions()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	{
		vers := make([]string, 0, len(local))
		for i := range local {
			vers = append(vers, local[i].Version)
		}
		fmt.Printf("[ %s ]\n", strings.Join(vers, ", "))
	}

	fmt.Printf("listing latest verions\n")
	remote, err := GetLatestVersions(flags.NumVersions)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	{
		vers := make([]string, 0, len(remote))
		for i := range remote {
			vers = append(vers, remote[i].Version)
		}
		fmt.Printf("[ %s ]\n", strings.Join(vers, ", "))
	}

	diff := CalcDiff(remote, local)

	for _, v := range diff.Remove {
		fmt.Printf("removing %s\n", v.Version)
		if err := RemoveVersion(v); err != nil {
			fmt.Printf("error: %v\n", err)
			continue
		}
	}

	for _, v := range diff.Install {
		fmt.Printf("installing %s\n", v.Version)
		if err := InstallVerion(v); err != nil {
			fmt.Printf("error: %v\n", err)
			continue
		}
	}

	fmt.Printf("done\n")
}

type Diff struct {
	Remove  []LocalVersion
	Install []RemoteVersion
}

func CalcDiff(remotes []RemoteVersion, locals []LocalVersion) Diff {
	diff := Diff{
		Remove:  make([]LocalVersion, 0, len(locals)),
		Install: make([]RemoteVersion, 0, len(remotes)),
	}

	rmap := map[string]*RemoteVersion{}
	for i := range remotes {
		v := &remotes[i]
		rmap[v.Version] = v
	}
	lmap := map[string]*LocalVersion{}
	for i := range locals {
		v := &locals[i]
		lmap[v.Version] = v
	}

	for k, v := range rmap {
		if _, ok := lmap[k]; !ok {
			diff.Install = append(diff.Install, *v)
		}
	}
	for k, v := range lmap {
		if _, ok := rmap[k]; !ok {
			diff.Remove = append(diff.Remove, *v)
		}
	}

	return diff
}

type LocalVersion struct {
	Version string
	Major   int
	Minor   int
	Patch   int

	SDKDir         string
	ExecutablePath string
}

func GetLocalVersions() ([]LocalVersion, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	root := filepath.Join(home, "sdk")
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("can't read go sdk dir: %w", err)
	}

	var versions []LocalVersion
	for _, e := range entries {
		version := LocalVersion{
			Version:        e.Name(),
			SDKDir:         filepath.Join(root, e.Name()),
			ExecutablePath: filepath.Join(home, "go", "bin", e.Name()),
		}

		match := regexpGoVersion.FindStringSubmatch(version.Version)
		if match == nil {
			continue
		}
		version.Major, err = strconv.Atoi(match[1])
		if err != nil {
			continue
		}
		version.Minor, err = strconv.Atoi(match[2])
		if err != nil {
			continue
		}
		version.Patch, err = strconv.Atoi(match[3])
		if err != nil {
			continue
		}

		versions = append(versions, version)
	}

	sort.Sort(byLocalVersion(versions))

	return versions, nil
}

type RemoteVersion struct {
	Version string       `json:"version,omitempty"`
	Major   int          `json:"major,omitempty"`
	Minor   int          `json:"minor,omitempty"`
	Patch   int          `json:"patch,omitempty"`
	Stable  bool         `json:"stable,omitempty"`
	Files   []RemoteFile `json:"files,omitempty"`
}

type RemoteFile struct {
	Filename string `json:"filename,omitempty"`
	OS       string `json:"os,omitempty"`
	Arch     string `json:"arch,omitempty"`
	Hash     string `json:"sha256,omitempty"`
	Size     int    `json:"size,omitempty"`
	Kind     string `json:"kind,omitempty"`
}

func GetLatestVersions(count int) ([]RemoteVersion, error) {
	if count < 1 {
		return nil, nil
	}
	uri := "https://go.dev/dl/?mode=json"
	if count > 2 {
		uri += "&include=all"
	}
	res, err := http.Get(uri)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	dec := json.NewDecoder(res.Body)

	t, err := dec.Token()
	if err != nil || t != json.Delim('[') {
		return nil, fmt.Errorf("expected array: %w", err)
	}

	var versions []RemoteVersion
	minors := map[uint64]*RemoteVersion{}

	for dec.More() {
		var version RemoteVersion
		if err := dec.Decode(&version); err != nil {
			return nil, fmt.Errorf("can't decode version: %w", err)
		}

		match := regexpGoVersion.FindStringSubmatch(version.Version)
		if match == nil {
			continue
		}
		version.Major, err = strconv.Atoi(match[1])
		if err != nil {
			continue
		}
		version.Minor, err = strconv.Atoi(match[2])
		if err != nil {
			continue
		}
		version.Patch, err = strconv.Atoi(match[3])
		if err != nil {
			continue
		}

		key := uint64(version.Major<<32) | uint64(version.Minor)
		if _, ok := minors[key]; !ok {
			minors[key] = &version
			versions = append(versions, version)
			if len(versions) == count {
				break
			}
		}
	}

	return versions, nil
}

func InstallVerion(v RemoteVersion) error {
	var filename string
	for _, f := range v.Files {
		if f.OS == "linux" && f.Arch == "amd64" {
			filename = f.Filename
			break
		}
	}

	homeDir, _ := os.UserHomeDir()
	sdkDir := filepath.Join(homeDir, "sdk")
	if err := os.MkdirAll(sdkDir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(sdkDir, filename)); errors.Is(err, os.ErrNotExist) {
		cmd := exec.Command("curl", "-LO", "https://go.dev/dl/"+filename)
		cmd.Dir = sdkDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("download go tarball: %w", err)
		}
	}

	cmd := exec.Command("tar", "-xzf", filename)
	cmd.Dir = sdkDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("extract go tarball: %w", err)
	}

	if err := os.Rename(filepath.Join(sdkDir, "go"), filepath.Join(sdkDir, v.Version)); err != nil {
		return err
	}

	return nil
}

func RemoveVersion(v LocalVersion) error {
	if err := os.RemoveAll(v.SDKDir); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("can't remove sdk: %w", err)
		}
	}
	if err := os.Remove(v.ExecutablePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("can't remove executable: %w", err)
		}
	}
	return nil
}

var regexpGoVersion = regexp.MustCompile(`^go(\d+)\.(\d+)\.(\d+)$`)

type byLocalVersion []LocalVersion

func (s byLocalVersion) Len() int {
	return len(s)
}

func (s byLocalVersion) Less(i int, j int) bool {
	return s[i].Major > s[j].Major ||
		(s[i].Major == s[j].Major && (s[i].Minor > s[j].Minor ||
			(s[i].Minor == s[j].Minor && s[i].Patch > s[j].Patch)))
}

func (s byLocalVersion) Swap(i int, j int) {
	s[i], s[j] = s[j], s[i]
}

func ParseFlags(f any) {
	var helpMsg strings.Builder
	helpMsg.WriteString("godl Makes sure N latest go versions are installed in ~/sdk\n")
	helpMsg.WriteString("\nUsage:\n")
	helpMsg.WriteString("godl [flags]\n")

	t := reflect.TypeOf(f).Elem()
	v := reflect.ValueOf(f).Elem()
	if t.NumField() > 0 {
		helpMsg.WriteString("\nFlags:")
	}

	var flagNames, flagUsages []string
	for i := 0; i < v.NumField(); i++ {
		ft := t.Field(i)
		fv := v.Field(i).Addr().Interface()

		short := ft.Tag.Get("short")
		long := ft.Tag.Get("long")
		def := ft.Tag.Get("default")
		usage := ft.Tag.Get("usage")

		var fvariants []string
		if short != "" {
			fvariants = append(fvariants, "-"+short)
		}
		if long != "" {
			fvariants = append(fvariants, "--"+long)
		}
		flagName := strings.Join(fvariants, ", ")
		if def != "" {
			flagName += "[=" + def + "]"
		}
		flagNames = append(flagNames, flagName)
		flagUsages = append(flagUsages, usage)

		switch fv := fv.(type) {
		case *bool:
			{
				defv := false
				if def != "" {
					v, err := strconv.ParseBool(def)
					if err != nil {
						panic("invalid bool default: " + def)
					}
					defv = v
				}
				if short != "" {
					flag.BoolVar(fv, short, defv, usage)
				}
				if long != "" {
					flag.BoolVar(fv, long, defv, usage)
				}
			}
		case *string:
			if short != "" {
				flag.StringVar(fv, short, def, usage)
			}
			if long != "" {
				flag.StringVar(fv, long, def, usage)
			}
		case *int:
			{
				defv := 0
				if def != "" {
					v, err := strconv.Atoi(def)
					if err != nil {
						panic("invalid int default: " + def)
					}
					defv = v
				}
				if short != "" {
					flag.IntVar(fv, short, defv, usage)
				}
				if long != "" {
					flag.IntVar(fv, long, defv, usage)
				}
			}
		}
	}

	longestName := 0
	for i := range flagNames {
		if l := len(flagNames[i]); l > longestName {
			longestName = l
		}
	}
	for i := range flagNames {
		fnl := len(flagNames[i])
		helpMsg.WriteString("\n  ")
		helpMsg.WriteString(flagNames[i])
		helpMsg.WriteString(strings.Repeat(" ", longestName-fnl+2))
		helpMsg.WriteString(flagUsages[i])
	}

	flag.Usage = func() {
		fmt.Println(helpMsg.String())
	}

	flag.Parse()
}
