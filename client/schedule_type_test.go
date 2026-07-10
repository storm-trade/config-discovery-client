package client

import (
	"testing"

	"github.com/storm-trade/config-discovery-client/types"
	"github.com/test-go/testify/require"
)

func TestConfigDiscoveryScheduleHelpers(t *testing.T) {
	c := &configDiscovery{
		Schedules: map[string]*types.AssetSchedule{},
	}
	require.Equal(t, types.ScheduleTypeEffective, c.GetScheduleType("BTC"))
	require.True(t, c.IsScheduleEffective("BTC"))

	c.Schedules["SPCX"] = &types.AssetSchedule{
		ScheduleType: types.ScheduleTypeInfo,
	}
	require.Equal(t, types.ScheduleTypeInfo, c.GetScheduleType("SPCX"))
	require.False(t, c.IsScheduleEffective("SPCX"))

	c.Schedules["ETH"] = &types.AssetSchedule{}
	require.Equal(t, types.ScheduleTypeEffective, c.GetScheduleType("ETH"))
	require.True(t, c.IsScheduleEffective("ETH"))
}
