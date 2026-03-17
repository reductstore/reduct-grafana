package plugin

import (
	"testing"

	"github.com/reductstore/reductstore/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestClientOptionsFromSettings(t *testing.T) {
	options := clientOptionsFromSettings(&models.PluginSettings{
		VerifySSL:  true,
		CACertPath: "/tmp/ca.pem",
		Secrets: &models.SecretPluginSettings{
			ServerToken: "tok",
		},
	})

	assert.Equal(t, "tok", options.APIToken)
	assert.False(t, options.InsecureSkipVerify)
	assert.Equal(t, "/tmp/ca.pem", options.CACertPath)
}

func TestClientOptionsFromSettingsDisablesVerificationWhenRequested(t *testing.T) {
	options := clientOptionsFromSettings(&models.PluginSettings{
		VerifySSL: false,
		Secrets: &models.SecretPluginSettings{
			ServerToken: "tok",
		},
	})

	assert.True(t, options.InsecureSkipVerify)
}
