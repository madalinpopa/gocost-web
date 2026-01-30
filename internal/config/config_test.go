package config_test

import (
	"bytes"
	"log/slog"
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
				Currency:     "USD",
			},
			wantErr: false,
		},
		{
			name: "Space separated ALLOWED_HOSTS",
			envVars: map[string]string{
				"ALLOWED_HOSTS": "localhost example.com",
				"DOMAIN":        "gocost.ro",
			},
			want: &config.Config{
				Addr:         "0.0.0.0",
				Port:         4000,
				Dsn:          "data.sqlite",
				AllowedHosts: []string{"localhost", "example.com"},
				Domain:       "gocost.ro",
				Currency:     "USD",
			},
			wantErr: false,
		},
		{
			name: "Comma separated ALLOWED_HOSTS",
			envVars: map[string]string{
				"ALLOWED_HOSTS": "localhost,example.com",
				"DOMAIN":        "gocost.ro",
			},
			want: &config.Config{
				Addr:         "0.0.0.0",
				Port:         4000,
				Dsn:          "data.sqlite",
				AllowedHosts: []string{"localhost", "example.com"},
				Domain:       "gocost.ro",
				Currency:     "USD",
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
		{
			name: "Valid currency uses provided code",
			envVars: map[string]string{
				"ALLOWED_HOSTS": "localhost",
				"DOMAIN":        "gocost.ro",
				"CURRENCY":      "EUR",
			},
			want: &config.Config{
				Addr:         "0.0.0.0",
				Port:         4000,
				Dsn:          "data.sqlite",
				AllowedHosts: []string{"localhost"},
				Domain:       "gocost.ro",
				Currency:     "EUR",
			},
			wantErr: false,
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

func TestConfig_LoadEnvironments_InvalidCurrency_UsesDefaultAndLogs(t *testing.T) {
	type configWithBuffer struct {
		cfg    *config.Config
		buffer *bytes.Buffer
	}

	tests := []struct {
		name       string
		newConfig  func() configWithBuffer
		defaultLog *bytes.Buffer
	}{
		{
			name: "Uses default logger when nil",
			newConfig: func() configWithBuffer {
				return configWithBuffer{cfg: config.New()}
			},
			defaultLog: &bytes.Buffer{},
		},
		{
			name: "Uses provided logger when set",
			newConfig: func() configWithBuffer {
				buf := &bytes.Buffer{}
				logger := slog.New(slog.NewTextHandler(buf, nil))
				return configWithBuffer{
					cfg:    config.NewWithLogger(logger),
					buffer: buf,
				}
			},
			defaultLog: &bytes.Buffer{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			if err := os.Setenv("ALLOWED_HOSTS", "localhost"); err != nil {
				t.Fatalf("failed to set ALLOWED_HOSTS: %v", err)
			}
			if err := os.Setenv("DOMAIN", "gocost.ro"); err != nil {
				t.Fatalf("failed to set DOMAIN: %v", err)
			}
			if err := os.Setenv("CURRENCY", "NOT_A_CURRENCY"); err != nil {
				t.Fatalf("failed to set CURRENCY: %v", err)
			}

			originalDefault := slog.Default()
			defaultBuf := tt.defaultLog
			defaultLogger := slog.New(slog.NewTextHandler(defaultBuf, nil))
			slog.SetDefault(defaultLogger)
			t.Cleanup(func() {
				slog.SetDefault(originalDefault)
			})

			got := tt.newConfig()
			cfg := got.cfg.WithServerAddr("0.0.0.0", 4000).WithDatabaseDsn("data.sqlite")
			err := cfg.LoadEnvironments()
			if err != nil {
				t.Fatalf("LoadEnvironments() error = %v, want nil", err)
			}

			if cfg.Currency != "USD" {
				t.Fatalf("Currency = %s, want USD", cfg.Currency)
			}

			expectedLog := "invalid currency code; using default"
			if got.buffer != nil {
				if !bytes.Contains(got.buffer.Bytes(), []byte(expectedLog)) {
					t.Fatalf("expected custom logger to contain %q, got %q", expectedLog, got.buffer.String())
				}
				if bytes.Contains(defaultBuf.Bytes(), []byte(expectedLog)) {
					t.Fatalf("did not expect default logger to contain %q, got %q", expectedLog, defaultBuf.String())
				}
				return
			}

			if !bytes.Contains(defaultBuf.Bytes(), []byte(expectedLog)) {
				t.Fatalf("expected default logger to contain %q, got %q", expectedLog, defaultBuf.String())
			}
		})
	}
}
