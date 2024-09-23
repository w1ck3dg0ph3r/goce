package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Config stores configuration options for goce.
type Config struct {
	// Address and/or port to listen on.
	Listen string

	// TTL of the compilation result cache.
	CompilationCacheTTL time.Duration
	// TTL of the shared code.
	SharedCodeTTL time.Duration

	Compilers struct {
		// Search $PATH for go compilers.
		SearchGoPath bool
		// Search $HOME/sdk/go* for go compilers.
		SearchSDKPath bool

		// Paths of local go compiler executables.
		LocalCompilers []string

		// Add supported cross-compilation architectures.
		AdditionalArchitectures bool

		// Enable modules support.
		EnableModules bool
	}

	Cache struct {
		Enabled bool
	}
}

// Read reads configuration options from available sources.
func Read() (*Config, error) {
	viper.MustBindEnv("Listen", "GOCE_LISTEN")
	viper.MustBindEnv("CompilationCacheTTL", "GOCE_COMPILATION_CACHE_TTL")
	viper.MustBindEnv("SharedCodeTTL", "GOCE_SHARED_CODE_TTL")
	viper.MustBindEnv("Compilers.SearchGoPath", "GOCE_COMPILERS_SEARCH_GO_PATH")
	viper.MustBindEnv("Compilers.SearchSDKPath", "GOCE_COMPILERS_SEARCH_SDK_PATH")
	viper.MustBindEnv("Compilers.LocalCompilers", "GOCE_COMPILERS_LOCAL_COMPILERS")
	viper.MustBindEnv("Compilers.AdditionalArchitectures", "GOCE_COMPILERS_ADDITIONAL_ARCHITECTURES")
	viper.MustBindEnv("Compilers.EnableModules", "GOCE_COMPILERS_ENABLE_MODULES")
	viper.MustBindEnv("Cache.Enabled", "GOCE_CACHE_ENABLED")

	viper.SetDefault("Listen", ":9000")
	viper.SetDefault("CompilationCacheTTL", 2*time.Hour)
	viper.SetDefault("SharedCodeTTL", 24*time.Hour)
	viper.SetDefault("Compilers.SearchGoPath", true)
	viper.SetDefault("Compilers.SearchSDKPath", true)
	viper.SetDefault("Compilers.LocalCompilers", []string{})
	viper.SetDefault("Compilers.AdditionalArchitectures", true)
	viper.SetDefault("Compilers.EnableModules", true)
	viper.SetDefault("Cache.Enabled", true)

	home, err := os.UserHomeDir()
	if err != nil {
		home = "$HOME"
	}
	viper.SetConfigName("goce.toml")
	viper.SetConfigType("toml")
	viper.AddConfigPath("/etc/goce")
	viper.AddConfigPath(filepath.Join(home, ".config/goce"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.UnmarshalExact(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
