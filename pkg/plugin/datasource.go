package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	reductgo "github.com/reductstore/reduct-go"
	"github.com/reductstore/reductstore/pkg/models"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces - only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*ReductDatasource)(nil)
	_ backend.CheckHealthHandler    = (*ReductDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*ReductDatasource)(nil)
)

// NewDatasource creates a new datasource instance.
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	// opts, err := settings.HTTPClientOptions(ctx)
	// if err != nil {
	// 	return nil, fmt.Errorf("http client options: %w", err)
	// }

	// // Important: Reuse the same client for each query to avoid using all available connections on a host.

	// opts.ForwardHTTPHeaders = true

	// cl, err := httpclient.New(opts)
	// if err != nil {
	// 	return nil, fmt.Errorf("httpclient new: %w", err)
	// }
	// Get the URL and API token from JSON config
	pluginSettings, err := models.LoadPluginSettings(settings)
	if err != nil {
		return nil, fmt.Errorf("load plugin settings: %w", err)
	}
	// check both server url and server token are in the plugin settings
	fmt.Println("pluginSettings", pluginSettings)
	if pluginSettings.ServerURL == "" || pluginSettings.Secrets.ServerToken == "" {
		return nil, fmt.Errorf("server url and server token are required")
	}
	client := reductgo.NewClient(pluginSettings.ServerURL, reductgo.ClientOptions{
		APIToken:  pluginSettings.Secrets.ServerToken,
		VerifySSL: pluginSettings.VerifySSL,
	})
	_, err = client.IsLive(ctx)
	if err != nil {
		return nil, fmt.Errorf("check health: %w", err)
	}

	return &ReductDatasource{reductClient: client}, nil
}

// ReductDatasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type ReductDatasource struct {
	reductClient reductgo.Client
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *ReductDatasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *ReductDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

func (d *ReductDatasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse
	// Unmarshal the JSON into our queryModel.
	// create data frame response.
	// For an overview on data frames and how grafana handles them:
	// https://grafana.com/developers/plugin-tools/introduction/data-frames

	switch query.QueryType {
	case "getInfo":
		frame, err := d.reductClient.GetInfo(ctx)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("get info: %v", err.Error()))
		}
		// use a single frame for all the info
		infoFrame := data.NewFrame("Info",
			data.NewField("Version", nil, []string{frame.Version}),
			data.NewField("Bucket Count", nil, []int64{frame.BucketCount}),
			data.NewField("Usage", nil, []uint64{frame.Usage}),
			data.NewField("Uptime", nil, []uint64{frame.Uptime}),
			data.NewField("Oldest Record", nil, []uint64{frame.OldestRecord}),
			data.NewField("Latest Record", nil, []uint64{frame.LatestRecord}),
		)
		// check nil
		if frame.License != nil {
			infoFrame.Fields = append(infoFrame.Fields,
				data.NewField("Licensee", nil, []string{frame.License.Licensee}),
				data.NewField("Invoice", nil, []string{frame.License.Invoice}),
				data.NewField("Expiry Date", nil, []string{frame.License.ExpiryDate}),
				data.NewField("Plan", nil, []string{frame.License.Plan}),
				data.NewField("Device Number", nil, []int64{frame.License.DeviceNumber}),
				data.NewField("Disk Quota", nil, []int64{frame.License.DiskQuota}),
				data.NewField("Fingerprint", nil, []string{frame.License.Fingerprint}),
			)
		}
		if frame.Defaults.Bucket.MaxBlockSize != 0 {
			infoFrame.Fields = append(infoFrame.Fields, data.NewField("Max Block Size", nil, []int64{frame.Defaults.Bucket.MaxBlockSize}))
		}
		if frame.Defaults.Bucket.MaxBlockRecords != 0 {
			infoFrame.Fields = append(infoFrame.Fields, data.NewField("Max Block Records", nil, []int64{frame.Defaults.Bucket.MaxBlockRecords}))
		}
		if frame.Defaults.Bucket.QuotaType != "" {
			infoFrame.Fields = append(infoFrame.Fields, data.NewField("Quota Type", nil, []string{string(frame.Defaults.Bucket.QuotaType)}))
		}
		if frame.Defaults.Bucket.QuotaSize != 0 {
			infoFrame.Fields = append(infoFrame.Fields, data.NewField("Quota Size", nil, []int64{frame.Defaults.Bucket.QuotaSize}))
		}
		response.Frames = append(response.Frames, infoFrame)
	case "listTokens":
		frames, err := d.listTokens(ctx)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("list tokens: %v", err.Error()))
		}
		response.Frames = append(response.Frames, frames...)
	case "listBuckets":
		frames, err := d.listBuckets(ctx)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("list buckets: %v", err.Error()))
		}
		response.Frames = append(response.Frames, frames...)
	case "getBucketEntries":
		frames, err := d.getBucketEntries(ctx, pCtx, query)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("get entries: %v", err.Error()))
		}
		response.Frames = append(response.Frames, frames...)
	case "getBucketSetting":
		frame, err := d.getBucketSetting(ctx, pCtx, query)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("get bucket settings: %v", err.Error()))
		}
		response.Frames = append(response.Frames, frame)
	case "getReplicationTasks":
		frames, err := d.getReplicationTasks(ctx)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("get replication tasks: %v", err.Error()))
		}
		response.Frames = append(response.Frames, frames...)
	case "queryRecords":
		result, err := d.reductQuery(ctx, query)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("query: %v", err.Error()))
		}
		response.Frames = append(response.Frames, result...)
	default:
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("invalid query type: %s", query.QueryType))
	}

	return response
}

type reductQuery struct {
	Bucket  string                `json:"bucket"`
	Entry   string                `json:"entry"`
	Options reductgo.QueryOptions `json:"options"`
}

func (d *ReductDatasource) listTokens(ctx context.Context) ([]*data.Frame, error) {
	tokens, err := d.reductClient.GetTokens(ctx)
	if err != nil {
		return nil, err
	}
	frame := data.NewFrame("Tokens",
		data.NewField("Name", nil, []string{}),
		data.NewField("Created At", nil, []string{}),
		data.NewField("Is Provisioned", nil, []bool{}),
		data.NewField("Full Access", nil, []bool{}),
		data.NewField("Read", nil, []string{}),
		data.NewField("Write", nil, []string{}),
	)
	for _, token := range tokens {
		frame.Fields[0].Append(token.Name)
		frame.Fields[1].Append(token.CreatedAt)
		frame.Fields[2].Append(token.IsProvisioned)
		if token.Permissions != nil {
			frame.Fields[3].Append(token.Permissions.FullAccess)
			frame.Fields[4].Append(strings.Join(token.Permissions.Read, ","))
			frame.Fields[5].Append(strings.Join(token.Permissions.Write, ","))
		}
	}
	return []*data.Frame{frame}, nil
}

func (d *ReductDatasource) getReplicationTasks(ctx context.Context) ([]*data.Frame, error) {
	tasks, err := d.reductClient.GetReplicationTasks(ctx)
	if err != nil {
		return nil, err
	}
	frame := data.NewFrame("Replication Tasks",
		data.NewField("Name", nil, []string{}),
		data.NewField("Is Active", nil, []bool{}),
		data.NewField("Is Provisioned", nil, []bool{}),
		data.NewField("Pending Records", nil, []int64{}),
	)
	for _, task := range tasks {
		frame.Fields[0].Append(task.Name)
		frame.Fields[1].Append(task.IsActive)
		frame.Fields[2].Append(task.IsProvisioned)
		frame.Fields[3].Append(task.PendingRecords)
	}
	return []*data.Frame{frame}, nil
}

func (d *ReductDatasource) reductQuery(ctx context.Context, query backend.DataQuery) ([]*data.Frame, error) {
	var qm reductQuery
	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		return nil, err
	}
	bucket, err := d.reductClient.GetBucket(ctx, qm.Bucket)
	if err != nil {
		return nil, err
	}
	options := reductgo.QueryOptions{
		QueryType: reductgo.QueryTypeQuery,
	}
	if qm.Options.Start != 0 {
		options.Start = qm.Options.Start
	}
	if qm.Options.Stop != 0 {
		options.Stop = qm.Options.Stop
	}
	if qm.Options.When != nil {
		options.When = qm.Options.When
	}
	if qm.Options.Ext != nil {
		options.Ext = qm.Options.Ext
	}
	options.Strict = qm.Options.Strict
	options.Continuous = qm.Options.Continuous

	records, err := bucket.Query(ctx, qm.Entry, &options)
	if err != nil {
		return nil, err
	}

	recordFrame := data.NewFrame("Records",
		data.NewField("Data", nil, []json.RawMessage{}),
		data.NewField("Time", nil, []int64{}),
		data.NewField("Size", nil, []int64{}),
		data.NewField("Labels", nil, []json.RawMessage{}),
	)
	for record := range records.Records() {
		recordData, err := record.Read()
		if err != nil {
			return nil, err
		}
		recordFrame.Fields[0].Append(json.RawMessage(recordData))
		recordFrame.Fields[1].Append(record.Time())
		recordFrame.Fields[2].Append(record.Size())
		for key, value := range record.Labels() {
			recordFrame.Fields[3].Labels[key] = fmt.Sprintf("%v", value)
		}
		// json stringify the value
		jsonValue, err := json.Marshal(record.Labels())
		if err != nil {
			return nil, err
		}
		recordFrame.Fields[3].Append(json.RawMessage(jsonValue))
	}

	return []*data.Frame{recordFrame}, nil
}

func (d *ReductDatasource) listBuckets(ctx context.Context) ([]*data.Frame, error) {
	buckets, err := d.reductClient.GetBuckets(ctx)
	if err != nil {
		return nil, err
	}
	bucketFrame := data.NewFrame("Buckets",
		data.NewField("Name", nil, []string{}),
		data.NewField("Entry Count", nil, []int64{}),
		data.NewField("Size (bytes)", nil, []int64{}),
		data.NewField("Oldest Record (microseconds)", nil, []uint64{}),
		data.NewField("Latest Record (microseconds)", nil, []uint64{}),
		data.NewField("Provisioned", nil, []bool{}),
	)
	for _, bucket := range buckets {
		bucketFrame.Fields[0].Append(bucket.Name)
		bucketFrame.Fields[1].Append(bucket.EntryCount)
		bucketFrame.Fields[2].Append(bucket.Size)
		bucketFrame.Fields[3].Append(bucket.OldestRecord)
		bucketFrame.Fields[4].Append(bucket.LatestRecord)
		bucketFrame.Fields[5].Append(bucket.IsProvisioned)
	}
	return []*data.Frame{bucketFrame}, nil
}

type getEntriesQuery struct {
	Bucket string `json:"bucket"`
}

type getBucketSettingsQuery struct {
	Bucket string `json:"bucket"`
}

func (d *ReductDatasource) getBucketSetting(ctx context.Context, _ backend.PluginContext, query backend.DataQuery) (*data.Frame, error) {
	var qm getBucketSettingsQuery
	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		return nil, err
	}
	bucket, err := d.reductClient.GetBucket(ctx, qm.Bucket)
	if err != nil {
		return nil, err
	}
	bucketSetting, err := bucket.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	frame := data.NewFrame("Bucket Settings",
		data.NewField("Max Block Size (bytes)", nil, []int64{bucketSetting.MaxBlockSize}),
		data.NewField("Max Block Records", nil, []int64{bucketSetting.MaxBlockRecords}),
		data.NewField("Quota Type", nil, []string{string(bucketSetting.QuotaType)}),
		data.NewField("Quota Size (bytes)", nil, []int64{bucketSetting.QuotaSize}),
	)
	return frame, nil
}
func (d *ReductDatasource) getBucketEntries(ctx context.Context, _ backend.PluginContext, query backend.DataQuery) ([]*data.Frame, error) {
	var qm getEntriesQuery
	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		return nil, err
	}
	bucket, err := d.reductClient.GetBucket(ctx, qm.Bucket)
	if err != nil {
		return nil, err
	}

	entries, err := bucket.GetEntries(ctx)
	if err != nil {
		return nil, err
	}
	entryFrame := data.NewFrame("Entries",
		data.NewField("Name", nil, []string{}),
		data.NewField("Size (bytes)", nil, []int64{}),
		data.NewField("Block Count", nil, []int64{}),
		data.NewField("Record Count", nil, []int64{}),
		data.NewField("Oldest Record (microseconds)", nil, []int64{}),
		data.NewField("Latest Record (microseconds)", nil, []int64{}),
	)
	for _, entry := range entries {
		entryFrame.Fields[0].Append(entry.Name)
		entryFrame.Fields[1].Append(entry.Size)
		entryFrame.Fields[2].Append(entry.BlockCount)
		entryFrame.Fields[3].Append(entry.RecordCount)
		entryFrame.Fields[4].Append(entry.OldestRecord)
		entryFrame.Fields[5].Append(entry.LatestRecord)
	}
	return []*data.Frame{entryFrame}, nil
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

	if pluginSettings.Secrets.ServerToken == "" {
		res.Status = backend.HealthStatusError
		res.Message = "Server token is missing"
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
