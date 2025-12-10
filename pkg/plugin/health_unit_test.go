package plugin

import (
	"context"
	"errors"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	reductgo "github.com/reductstore/reduct-go"
	"github.com/reductstore/reduct-go/model"
	"github.com/stretchr/testify/assert"
)

type stubClient struct {
	liveErr error
	infoErr error
}

func (s stubClient) GetInfo(ctx context.Context) (model.ServerInfo, error) {
	return model.ServerInfo{}, s.infoErr
}
func (s stubClient) IsLive(ctx context.Context) (bool, error) { return true, s.liveErr }
func (s stubClient) GetBuckets(ctx context.Context) ([]model.BucketInfo, error) {
	return nil, nil
}
func (s stubClient) CreateBucket(ctx context.Context, name string, settings *model.BucketSetting) (reductgo.Bucket, error) {
	return reductgo.Bucket{}, nil
}
func (s stubClient) CreateOrGetBucket(ctx context.Context, name string, settings *model.BucketSetting) (reductgo.Bucket, error) {
	return reductgo.Bucket{}, nil
}
func (s stubClient) GetBucket(ctx context.Context, name string) (reductgo.Bucket, error) {
	return reductgo.Bucket{}, nil
}
func (s stubClient) CheckBucketExists(ctx context.Context, name string) (bool, error) {
	return false, nil
}
func (s stubClient) RemoveBucket(ctx context.Context, name string) error  { return nil }
func (s stubClient) GetTokens(ctx context.Context) ([]model.Token, error) { return nil, nil }
func (s stubClient) GetToken(ctx context.Context, name string) (model.Token, error) {
	return model.Token{}, nil
}
func (s stubClient) CreateToken(ctx context.Context, name string, permissions model.TokenPermissions) (string, error) {
	return "", nil
}
func (s stubClient) RemoveToken(ctx context.Context, name string) error { return nil }
func (s stubClient) GetCurrentToken(ctx context.Context) (model.Token, error) {
	return model.Token{}, nil
}
func (s stubClient) GetReplicationTasks(ctx context.Context) ([]model.ReplicationInfo, error) {
	return nil, nil
}
func (s stubClient) GetReplicationTask(ctx context.Context, name string) (model.FullReplicationInfo, error) {
	return model.FullReplicationInfo{}, nil
}
func (s stubClient) CreateReplicationTask(ctx context.Context, name string, task model.ReplicationSettings) error {
	return nil
}
func (s stubClient) UpdateReplicationTask(ctx context.Context, name string, task model.ReplicationSettings) error {
	return nil
}
func (s stubClient) RemoveReplicationTask(ctx context.Context, name string) error { return nil }

func newCheckHealthRequest(jsonData string) *backend.CheckHealthRequest {
	return &backend.CheckHealthRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				JSONData:                []byte(jsonData),
				DecryptedSecureJSONData: map[string]string{"serverToken": "tok"},
			},
		},
	}
}

func TestCheckHealth(t *testing.T) {
	ds := &ReductDatasource{}

	t.Run("fails to load settings", func(t *testing.T) {
		result, err := ds.CheckHealth(context.Background(), newCheckHealthRequest(`{invalid`))
		assert.NoError(t, err)
		assert.Equal(t, backend.HealthStatusError, result.Status)
		assert.Equal(t, "Unable to load settings", result.Message)
	})

	t.Run("missing server URL", func(t *testing.T) {
		result, err := ds.CheckHealth(context.Background(), newCheckHealthRequest(`{}`))
		assert.NoError(t, err)
		assert.Equal(t, backend.HealthStatusError, result.Status)
		assert.Equal(t, "Server URL is missing", result.Message)
	})

	t.Run("unreachable server", func(t *testing.T) {
		orig := newReductClient
		newReductClient = func(url string, options reductgo.ClientOptions) reductgo.Client {
			return stubClient{liveErr: errors.New("boom")}
		}
		defer func() { newReductClient = orig }()

		result, err := ds.CheckHealth(context.Background(), newCheckHealthRequest(`{"serverURL":"http://x"}`))
		assert.NoError(t, err)
		assert.Equal(t, backend.HealthStatusError, result.Status)
		assert.Equal(t, "Unable to connect to server", result.Message)
	})

	t.Run("auth failure", func(t *testing.T) {
		orig := newReductClient
		newReductClient = func(url string, options reductgo.ClientOptions) reductgo.Client {
			return stubClient{infoErr: errors.New("auth fail")}
		}
		defer func() { newReductClient = orig }()

		result, err := ds.CheckHealth(context.Background(), newCheckHealthRequest(`{"serverURL":"http://x"}`))
		assert.NoError(t, err)
		assert.Equal(t, backend.HealthStatusError, result.Status)
		assert.Equal(t, "Authentication failed or server error", result.Message)
	})

	t.Run("happy path", func(t *testing.T) {
		orig := newReductClient
		newReductClient = func(url string, options reductgo.ClientOptions) reductgo.Client {
			return stubClient{}
		}
		defer func() { newReductClient = orig }()

		result, err := ds.CheckHealth(context.Background(), newCheckHealthRequest(`{"serverURL":"http://x"}`))
		assert.NoError(t, err)
		assert.Equal(t, backend.HealthStatusOk, result.Status)
		assert.Equal(t, "Data source is working", result.Message)
	})
}
