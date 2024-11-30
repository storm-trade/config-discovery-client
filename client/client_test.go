package client_test

import (
	"github.com/storm-trade/config-discovery-client/client"
	"github.com/test-go/testify/require"
	"testing"
)

func TestConfigDiscoveryImpl_GetAssetByName(t *testing.T) {
	cDiscovery := client.New("http://localhost:55824/config")

	err := cDiscovery.ListenUpdates()
	require.Nil(t, err)

	assetInfo := cDiscovery.GetAssetByName("LTC")

	require.Equal(t, assetInfo.Name, "LTC")
	require.Equal(t, assetInfo.Index, 11)
}