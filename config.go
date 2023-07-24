package main

import (
	"time"

	"github.com/joho/godotenv"
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
	}
}

// ReadConfig reads configuration options from available sources.
func ReadConfig() (*Config, error) {
	_ = godotenv.Load()

	viper.BindEnv("Listen", "GOCE_LISTEN")
	viper.BindEnv("CompilationCacheTTL", "GOCE_COMPILATION_CACHE_TTL")
	viper.BindEnv("SharedCodeTTL", "GOCE_SHARED_CODE_TTL")
	viper.BindEnv("Compilers.SearchGoPath", "GOCE_COMPILERS_SEARCH_GO_PATH")
	viper.BindEnv("Compilers.SearchSDKPath", "GOCE_COMPILERS_SEARCH_SDK_PATH")
	viper.BindEnv("Compilers.LocalCompilers", "GOCE_COMPILERS_LOCAL_COMPILERS")
	viper.BindEnv("Compilers.AdditionalArchitectures", "GOCE_COMPILERS_ADDITIONAL_ARCHITECTURES")

	viper.SetDefault("Listen", ":9000")
	viper.SetDefault("CompilationCacheTTL", 2*time.Hour)
	viper.SetDefault("SharedCodeTTL", 24*time.Hour)
	viper.SetDefault("Compilers.SearchGoPath", true)
	viper.SetDefault("Compilers.SearchSDKPath", false)
	viper.SetDefault("Compilers.LocalCompilers", []string{})
	viper.SetDefault("Compilers.AdditionalArchitectures", true)

	viper.SetConfigName("goce")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/goce")
	viper.AddConfigPath("$HOME/.config/goce")

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
