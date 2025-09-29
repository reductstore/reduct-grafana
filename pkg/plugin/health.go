package plugin

import (
	"context"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	reductgo "github.com/reductstore/reduct-go"
	"github.com/reductstore/reductstore/pkg/models"
)

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

	client := reductgo.NewClient(pluginSettings.ServerURL, reductgo.ClientOptions{
		APIToken:  pluginSettings.Secrets.ServerToken,
		VerifySSL: pluginSettings.VerifySSL,
	})
	_, err = client.IsLive(ctx)
	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Unable to connect to server"
		return res, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}
