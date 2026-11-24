package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIBaseURL string `json:"api_base_url"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	AuthToken  string `json:"auth_token"`
	UserEmail  string `json:"user_email"`
}

func Dir() (string, error) {
	if v := os.Getenv("DEVTRAIL_HOME"); v != "" {
		return v, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "devtrail"), nil
}

func DBPath() (string, error) {
	d, err := Dir()
	if err != nil { return "", err }
	return filepath.Join(d, "devtrail.db"), nil
}

func Path() (string, error) {
	d, err := Dir()
	if err != nil { return "", err }
	return filepath.Join(d, "config.json"), nil
}

func Load() (Config, error) {
	p, err := Path()
	if err != nil { return Config{}, err }
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{APIBaseURL: "http://localhost:8080"}, nil
		}
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil { return Config{}, err }
	if c.APIBaseURL == "" { c.APIBaseURL = "http://localhost:8080" }
	return c, nil
}

func Save(c Config) error {
	d, err := Dir()
	if err != nil { return err }
	if err := os.MkdirAll(d, 0o755); err != nil { return err }
	p, err := Path()
	if err != nil { return err }
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil { return err }
	return os.WriteFile(p, b, 0o600)
}
