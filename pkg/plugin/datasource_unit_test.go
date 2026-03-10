package plugin

import (
	"context"
	"errors"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	reductgo "github.com/reductstore/reduct-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDatasourceSettings(jsonData string) backend.DataSourceInstanceSettings {
	return backend.DataSourceInstanceSettings{
		JSONData:                []byte(jsonData),
		DecryptedSecureJSONData: map[string]string{"serverToken": "tok"},
	}
}

func TestNewDatasource(t *testing.T) {
	t.Run("unsupported server version", func(t *testing.T) {
		orig := newReductClient
		newReductClient = func(url string, options reductgo.ClientOptions) reductgo.Client {
			return stubClient{version: "1.17.0"}
		}
		defer func() { newReductClient = orig }()

		instance, err := NewDatasource(context.Background(), newDatasourceSettings(`{"serverURL":"http://x"}`))
		require.Error(t, err)
		assert.Nil(t, instance)
		assert.Equal(t, "ReductStore 1.17.0 is not supported. Upgrade ReductStore to v1.18.0 or higher", err.Error())
	})

	t.Run("server info failure", func(t *testing.T) {
		orig := newReductClient
		newReductClient = func(url string, options reductgo.ClientOptions) reductgo.Client {
			return stubClient{infoErr: errors.New("auth fail")}
		}
		defer func() { newReductClient = orig }()

		instance, err := NewDatasource(context.Background(), newDatasourceSettings(`{"serverURL":"http://x"}`))
		require.Error(t, err)
		assert.Nil(t, instance)
		assert.Equal(t, "Authentication failed or server error", err.Error())
	})

	t.Run("supported server version", func(t *testing.T) {
		orig := newReductClient
		newReductClient = func(url string, options reductgo.ClientOptions) reductgo.Client {
			return stubClient{version: "1.18.0"}
		}
		defer func() { newReductClient = orig }()

		instance, err := NewDatasource(context.Background(), newDatasourceSettings(`{"serverURL":"http://x"}`))
		require.NoError(t, err)
		require.NotNil(t, instance)
	})
}
