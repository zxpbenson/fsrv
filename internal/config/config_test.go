package config

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		check   func(*testing.T, *Config)
	}{
		{
			name:    "default values",
			args:    []string{},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.Port != "8080" {
					t.Errorf("expected port 8080, got %s", cfg.Port)
				}
				if cfg.DelAble != false {
					t.Errorf("expected DelAble false, got %v", cfg.DelAble)
				}
				if cfg.Store != "./store" {
					t.Errorf("expected store ./store, got %s", cfg.Store)
				}
				if cfg.Max != 32 {
					t.Errorf("expected max 32, got %d", cfg.Max)
				}
			},
		},
		{
			name:    "custom port",
			args:    []string{"-p", "3000"},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.Port != "3000" {
					t.Errorf("expected port 3000, got %s", cfg.Port)
				}
			},
		},
		{
			name:    "enable delete",
			args:    []string{"-d"},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.DelAble != true {
					t.Errorf("expected DelAble true, got %v", cfg.DelAble)
				}
			},
		},
		{
			name:    "custom store",
			args:    []string{"-s", "/tmp/store"},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.Store != "/tmp/store" {
					t.Errorf("expected store /tmp/store, got %s", cfg.Store)
				}
			},
		},
		{
			name:    "custom max size",
			args:    []string{"-m", "33"},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.Max != 33 {
					t.Errorf("expected max 33, got %d", cfg.Max)
				}
			},
		},
		{
			name:    "custom hostname",
			args:    []string{"-n", "test-server"},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.Hostname != "test-server" {
					t.Errorf("expected hostname test-server, got %s", cfg.Hostname)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

// ExampleConfig_Parse demonstrates how to parse configuration
func ExampleConfig_Parse() {
	// Note: This example cannot be run as-is because it requires command-line flags
	// In real usage, you would call Parse() which reads from os.Args
	// cfg, err := config.Parse()
	// if err != nil {
	//     log.Fatal(err)
	// }
	// fmt.Printf("Server will listen on port %s\n", cfg.Port)
	// Output: Server will listen on port 8080
}
