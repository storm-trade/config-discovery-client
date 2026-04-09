package client_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/storm-trade/config-discovery-client/client"
	"github.com/test-go/testify/require"
)

func TestConfigDiscoveryImpl_GetAssetByName(t *testing.T) {
	cDiscovery, err := client.New("https://api.stage.stormtrade.dev/api/config")

	err = cDiscovery.ListenUpdates()
	require.Nil(t, err)

	assetInfo := cDiscovery.GetAssetByName("LTC")

	require.Equal(t, assetInfo.Name, "LTC")
	require.Equal(t, assetInfo.Index, 11)

	for range 10 {
		vpi, ok := cDiscovery.GetVPIParamsAtTimestamp("FARTCOIN", time.Now().UnixMilli()-1000)
		if !ok {
			fmt.Println("no vpi params found")
			continue
		}

		fmt.Println("vpi", vpi.MarketDepthShort, vpi.MarketDepthLong, vpi.Spread, vpi.K, vpi.Timestamp)
	}
}
