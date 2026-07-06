package client

import (
	"testing"

	"github.com/storm-trade/config-discovery-client/types"
	"github.com/test-go/testify/require"
)

func TestConfigDiscoveryScheduleHelpers(t *testing.T) {
	c := &configDiscovery{}
	require.Equal(t, types.ScheduleTypeEffective, c.GetScheduleType())
	require.True(t, c.IsScheduleEffective())

	c.ScheduleType = types.ScheduleTypeInfo
	require.Equal(t, types.ScheduleTypeInfo, c.GetScheduleType())
	require.False(t, c.IsScheduleEffective())
}
