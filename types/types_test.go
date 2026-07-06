package types

import (
	"encoding/json"
	"testing"

	"github.com/test-go/testify/require"
)

func TestAssetsScheduleDefaultTypeIsEffective(t *testing.T) {
	var schedules AssetsSchedule

	err := json.Unmarshal([]byte(`{"schedules":{"BTC":{"schedule":"0 0 * * 1-5","holidays":"[]"}}}`), &schedules)
	require.NoError(t, err)
	require.Equal(t, ScheduleTypeEffective, schedules.GetScheduleType())
	require.True(t, schedules.IsEffective())
}

func TestAssetsScheduleInfoTypeParsing(t *testing.T) {
	var schedules AssetsSchedule

	err := json.Unmarshal([]byte(`{"scheduleType":"info","schedules":{"BTC":{"schedule":"0 0 * * 1-5","holidays":"[]"}}}`), &schedules)
	require.NoError(t, err)
	require.Equal(t, ScheduleTypeInfo, schedules.GetScheduleType())
	require.False(t, schedules.IsEffective())
}

func TestScheduleTypeIsEffective(t *testing.T) {
	require.True(t, ScheduleType("").IsEffective())
	require.True(t, ScheduleTypeEffective.IsEffective())
	require.False(t, ScheduleTypeInfo.IsEffective())
	require.False(t, ScheduleType("unknown").IsEffective())
}
