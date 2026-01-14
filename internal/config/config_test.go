package config_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/madalinpopa/gocost-web/internal/config"
)

func TestNew(t *testing.T) {
	cfg := config.New().WithServerAddr("0.0.0.0", 4000).WithDatabaseDsn("data.sqlite")

	if cfg.Addr != "0.0.0.0" {
		t.Errorf("Expected default Addr to be '0.0.0.0', got %s", cfg.Addr)
	}

	if cfg.Port != 4000 {
		t.Errorf("Expected default Port to be 4000, got %d", cfg.Port)
	}

	if cfg.Dsn != "data.sqlite" {
		t.Errorf("Expected default Dsn to be 'data.sqlite', got %s", cfg.Dsn)
	}
}

func TestConfig_LoadEnvironments(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    *config.Config
		wantErr bool
	}{
		{
			name: "Valid environment variables",
			envVars: map[string]string{
				"ALLOWED_HOSTS": "localhost",
				"DOMAIN":        "gocost.ro",
			},
			want: &config.Config{
				Addr:         "0.0.0.0",
				Port:         4000,
				Dsn:          "data.sqlite",
				AllowedHosts: []string{"localhost"},
				Domain:       "gocost.ro",
				Currency:     "$",
			},
			wantErr: false,
		},
		{
			name: "Missing ALLOWED_HOSTS",
			envVars: map[string]string{
				"ADMIN_EMAIL":    "admin@example.com",
				"ADMIN_FULLNAME": "Test Admin",
				"ADMIN_USERNAME": "admin",
				"ADMIN_PASSWORD": "secret123",
				"DOMAIN":         "gocost.ro",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Missing DOMAIN",
			envVars: map[string]string{
				"ADMIN_EMAIL":    "admin@example.com",
				"ADMIN_FULLNAME": "Test Admin",
				"ADMIN_USERNAME": "admin",
				"ADMIN_PASSWORD": "secret123",
				"ALLOWED_HOSTS":  "localhost",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "All environment variables missing",
			envVars: map[string]string{},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()

			// Set environment variables for the test
			for k, v := range tt.envVars {
				err := os.Setenv(k, v)
				if err != nil {
					t.Fatalf("failed to set environment variable %s: %v", k, err)
				}
			}

			// CreateUser a fresh config instance for each test
			cfg := config.New().WithServerAddr("0.0.0.0", 4000).WithDatabaseDsn("data.sqlite")

			err := cfg.LoadEnvironments()

			// Test apperror cases
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadEnvironments() apperror = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Test successful case
			if !tt.wantErr {
				if !reflect.DeepEqual(cfg, tt.want) {
					t.Errorf("LoadEnvironments() got = %+v, want %+v", cfg, tt.want)
				}
			}
		})
	}
}
