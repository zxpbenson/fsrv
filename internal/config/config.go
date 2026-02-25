package config

import (
	"flag"
	"fmt"
	"os"

	"fsrv/internal/util"
)

// Config holds the application configuration
type Config struct {
	Port     string
	DelAble  bool
	Hostname string
	Store    string
	Max      int64
}

// Parse parses command line arguments and returns the configuration
func Parse(args []string) (*Config, error) {
	cfg := &Config{}

	// Get default hostname
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	// Create a new FlagSet to avoid global state pollution and enable better testing
	fs := flag.NewFlagSet("fsrv", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage of %s:\n", fs.Name())
		fmt.Fprintf(fs.Output(), "  A simple HTTP file server with upload, download, and delete capabilities.\n\n")
		fmt.Fprintf(fs.Output(), "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(fs.Output(), "\nExamples:\n")
		fmt.Fprintf(fs.Output(), "  %s -p 8081\n", fs.Name())
		fmt.Fprintf(fs.Output(), "  %s -s /tmp/files -d\n", fs.Name())
	}

	// Parse command line flags
	fs.StringVar(&cfg.Port, "p", "8080", "Specify the port to listen on")
	fs.BoolVar(&cfg.DelAble, "d", false, "Enable delete file by UI")
	fs.StringVar(&cfg.Store, "s", "./store", "Specify the directory to store files")
	fs.StringVar(&cfg.Hostname, "n", hostname, "Specify the server name, default hostname")
	fs.Int64Var(&cfg.Max, "m", 32, "Max file size to upload, power of 2 (e.g., 32 means 1<<32=4GB)")

	// Parse arguments
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	// Print configuration
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Port: %s\n", cfg.Port)
	fmt.Printf("  Store: %s\n", cfg.Store)
	fmt.Printf("  Hostname: %s\n", cfg.Hostname)
	fmt.Printf("  Delete enabled: %t\n", cfg.DelAble)
	fmt.Printf("  Max file size: %d -> %s\n", cfg.Max, util.HumanReadableSize(1<<cfg.Max))

	return cfg, nil
}
