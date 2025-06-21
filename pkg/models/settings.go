package models

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type PluginSettings struct {
	Path      string                `json:"path"`
	ServerURL string                `json:"serverURL"`
	VerifySSL bool                  `json:"verifySSL"`
	Secrets   *SecretPluginSettings `json:"-"`
}

type SecretPluginSettings struct {
	ServerToken string `json:"serverToken"`
}

func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	var settings PluginSettings
	err := json.Unmarshal(source.JSONData, &settings)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal PluginSettings json: %w", err)
	}

	settings.Secrets = &SecretPluginSettings{
		ServerToken: source.DecryptedSecureJSONData["serverToken"],
	}

	return &settings, nil
}

func LoadPluginSettingsFromMap(source map[string]string) (*PluginSettings, error) {
	return &PluginSettings{
		ServerURL: source["serverURL"],
		VerifySSL: source["verifySSL"] == "true",
		Secrets: &SecretPluginSettings{
			ServerToken: source["serverToken"],
		},
	}, nil
}
