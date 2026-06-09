package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"
)

// ErrHelpWanted is returned by loadConfiguration when the user requested the
// usage/help output (via -h/-help). It replaces conf.ErrHelpWanted from the
// previously used github.com/ardanlabs/conf dependency.
var ErrHelpWanted = errors.New("help requested")

// WebAPIConfiguration describes the web API configuration. This structure is
// populated by loadConfiguration from defaults, environment variables, command
// line flags and (optionally) a YAML configuration file.
type WebAPIConfiguration struct {
	Config struct {
		Path string
	}
	Web struct {
		APIHost         string
		DebugHost       string
		ReadTimeout     time.Duration
		WriteTimeout    time.Duration
		ShutdownTimeout time.Duration
	}
	Debug bool
	DB    struct {
		Filename string
	}
}

// envOr returns the value of the environment variable named key (prefixed with
// "CFG_") if set, otherwise it returns def.
func envOr(key, def string) string {
	if v, ok := os.LookupEnv("CFG_" + key); ok {
		return v
	}
	return def
}

// envBool reads a boolean environment variable, falling back to def.
func envBool(key string, def bool) bool {
	if v, ok := os.LookupEnv("CFG_" + key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

// envDuration reads a duration environment variable, falling back to def.
func envDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv("CFG_" + key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

// loadConfiguration creates a WebAPIConfiguration starting from defaults,
// environment variables and command line flags, then optionally overrides it
// with values from a YAML configuration file.
//
// Precedence (lowest to highest): built-in defaults < environment variables <
// command line flags < configuration file. Environment variables are prefixed
// with "CFG_" (e.g. CFG_WEB_APIHOST, CFG_DEBUG, CFG_DB_FILENAME).
func loadConfiguration() (WebAPIConfiguration, error) {
	var cfg WebAPIConfiguration

	// Defaults (overridable by environment variables).
	defConfigPath := envOr("CONFIG_PATH", "/conf/config.yml")
	defAPIHost := envOr("WEB_APIHOST", "0.0.0.0:3000")
	defDebugHost := envOr("WEB_DEBUGHOST", "0.0.0.0:4000")
	defReadTimeout := envDuration("WEB_READTIMEOUT", 5*time.Second)
	defWriteTimeout := envDuration("WEB_WRITETIMEOUT", 5*time.Second)
	defShutdownTimeout := envDuration("WEB_SHUTDOWNTIMEOUT", 5*time.Second)
	defDebug := envBool("DEBUG", false)
	defDBFilename := envOr("DB_FILENAME", "./database.db")

	// Command line flags override the environment-derived defaults.
	fs := flag.NewFlagSet("webapi", flag.ContinueOnError)
	fs.StringVar(&cfg.Config.Path, "config-path", defConfigPath, "path to the YAML configuration file")
	fs.StringVar(&cfg.Web.APIHost, "web-apihost", defAPIHost, "address (host:port) for the API server")
	fs.StringVar(&cfg.Web.DebugHost, "web-debughost", defDebugHost, "address (host:port) for the debug server")
	fs.DurationVar(&cfg.Web.ReadTimeout, "web-readtimeout", defReadTimeout, "HTTP read timeout")
	fs.DurationVar(&cfg.Web.WriteTimeout, "web-writetimeout", defWriteTimeout, "HTTP write timeout")
	fs.DurationVar(&cfg.Web.ShutdownTimeout, "web-shutdowntimeout", defShutdownTimeout, "graceful shutdown timeout")
	fs.BoolVar(&cfg.Debug, "debug", defDebug, "enable debug logging")
	fs.StringVar(&cfg.DB.Filename, "db-filename", defDBFilename, "path to the SQLite database file")

	if err := fs.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return cfg, ErrHelpWanted
		}
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	// Override values from YAML if specified and if it exists (useful in k8s/compose)
	fp, err := os.Open(cfg.Config.Path)
	if err != nil && !os.IsNotExist(err) {
		return cfg, fmt.Errorf("can't read the config file, while it exists: %w", err)
	} else if err == nil {
		yamlFile, err := io.ReadAll(fp)
		if err != nil {
			return cfg, fmt.Errorf("can't read config file: %w", err)
		}
		err = yaml.Unmarshal(yamlFile, &cfg)
		if err != nil {
			return cfg, fmt.Errorf("can't unmarshal config file: %w", err)
		}
		_ = fp.Close()
	}

	return cfg, nil
}
