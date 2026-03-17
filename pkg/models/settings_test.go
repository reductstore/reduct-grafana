package models

import (
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPluginSettingsDefaultsVerifySSLToTrue(t *testing.T) {
	settings, err := LoadPluginSettings(backend.DataSourceInstanceSettings{
		JSONData:                []byte(`{"serverURL":"https://x","caCertPath":"/tmp/ca.pem"}`),
		DecryptedSecureJSONData: map[string]string{"serverToken": "tok"},
	})
	require.NoError(t, err)

	assert.Equal(t, "https://x", settings.ServerURL)
	assert.True(t, settings.VerifySSL)
	assert.Equal(t, "/tmp/ca.pem", settings.CACertPath)
	assert.Equal(t, "tok", settings.Secrets.ServerToken)
}

func TestLoadPluginSettingsHonorsExplicitVerifySSLValue(t *testing.T) {
	settings, err := LoadPluginSettings(backend.DataSourceInstanceSettings{
		JSONData:                []byte(`{"serverURL":"https://x","verifySSL":false}`),
		DecryptedSecureJSONData: map[string]string{"serverToken": "tok"},
	})
	require.NoError(t, err)

	assert.False(t, settings.VerifySSL)
}

func TestLoadPluginSettingsFromMapDefaultsVerifySSLToTrue(t *testing.T) {
	settings, err := LoadPluginSettingsFromMap(map[string]string{
		"serverURL":   "https://x",
		"serverToken": "tok",
		"caCertPath":  "/tmp/ca.pem",
	})
	require.NoError(t, err)

	assert.True(t, settings.VerifySSL)
	assert.Equal(t, "/tmp/ca.pem", settings.CACertPath)
}

func TestLoadPluginSettingsFromMapHonorsExplicitVerifySSLValue(t *testing.T) {
	settings, err := LoadPluginSettingsFromMap(map[string]string{
		"serverURL":   "https://x",
		"serverToken": "tok",
		"verifySSL":   "false",
	})
	require.NoError(t, err)

	assert.False(t, settings.VerifySSL)
}
