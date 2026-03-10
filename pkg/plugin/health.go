package plugin

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	reductgo "github.com/reductstore/reduct-go"
	reductmodel "github.com/reductstore/reduct-go/model"
	"github.com/reductstore/reductstore/pkg/models"
)

var newReductClient = reductgo.NewClient

const minimumReductStoreVersion = "v1.18.0"

func validateServerVersion(version string) error {
	serverVersion, err := reductmodel.ParseVersion(version)
	if err != nil {
		return fmt.Errorf("unable to determine ReductStore version: %w", err)
	}

	minimumVersion, err := reductmodel.ParseVersion(minimumReductStoreVersion)
	if err != nil {
		return fmt.Errorf("invalid minimum ReductStore version: %w", err)
	}

	if serverVersion.Major < minimumVersion.Major ||
		(serverVersion.Major == minimumVersion.Major && serverVersion.Minor < minimumVersion.Minor) {
		return fmt.Errorf(
			"ReductStore %s is not supported. Upgrade ReductStore to %s or higher",
			version,
			minimumReductStoreVersion,
		)
	}

	return nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *ReductDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	res := &backend.CheckHealthResult{}
	// url is not in secured json data, its in json data
	pluginSettings, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Unable to load settings"
		return res, nil
	}

	// check for server url
	if pluginSettings.ServerURL == "" {
		res.Status = backend.HealthStatusError
		res.Message = "Server URL is missing"
		return res, nil
	}
	// check for server token

	client := newReductClient(pluginSettings.ServerURL, reductgo.ClientOptions{
		APIToken:  pluginSettings.Secrets.ServerToken,
		VerifySSL: pluginSettings.VerifySSL,
	})
	_, err = client.IsLive(ctx)
	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Unable to connect to server"
		return res, nil
	}

	// Test authentication by trying to get server info
	serverInfo, err := client.GetInfo(ctx)
	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Authentication failed or server error"
		return res, nil
	}

	if err := validateServerVersion(serverInfo.Version); err != nil {
		res.Status = backend.HealthStatusError
		res.Message = err.Error()
		return res, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}
