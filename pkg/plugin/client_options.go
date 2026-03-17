package plugin

import (
	reductgo "github.com/reductstore/reduct-go"
	"github.com/reductstore/reductstore/pkg/models"
)

func clientOptionsFromSettings(settings *models.PluginSettings) reductgo.ClientOptions {
	return reductgo.ClientOptions{
		APIToken:           settings.Secrets.ServerToken,
		InsecureSkipVerify: !settings.VerifySSL,
		CACertPath:         settings.CACertPath,
	}
}
