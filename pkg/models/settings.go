package models

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type PluginSettings struct {
	ServerURL  string                `json:"serverURL"`
	VerifySSL  bool                  `json:"verifySSL"`
	CACertPath string                `json:"caCertPath"`
	Secrets    *SecretPluginSettings `json:"-"`
}

type SecretPluginSettings struct {
	ServerToken string `json:"serverToken"`
}

func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	var raw struct {
		ServerURL  string `json:"serverURL"`
		VerifySSL  *bool  `json:"verifySSL"`
		CACertPath string `json:"caCertPath"`
	}

	err := json.Unmarshal(source.JSONData, &raw)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal PluginSettings json: %w", err)
	}

	settings := PluginSettings{
		ServerURL:  raw.ServerURL,
		VerifySSL:  true,
		CACertPath: raw.CACertPath,
	}
	if raw.VerifySSL != nil {
		settings.VerifySSL = *raw.VerifySSL
	}

	settings.Secrets = &SecretPluginSettings{
		ServerToken: source.DecryptedSecureJSONData["serverToken"],
	}

	return &settings, nil
}

func LoadPluginSettingsFromMap(source map[string]string) (*PluginSettings, error) {
	verifySSL := true
	if value, ok := source["verifySSL"]; ok && value != "" {
		verifySSL = value == "true"
	}

	return &PluginSettings{
		ServerURL:  source["serverURL"],
		VerifySSL:  verifySSL,
		CACertPath: source["caCertPath"],
		Secrets: &SecretPluginSettings{
			ServerToken: source["serverToken"],
		},
	}, nil
}
