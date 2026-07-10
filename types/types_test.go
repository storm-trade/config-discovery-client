package types

import (
	"encoding/json"
	"testing"

	"github.com/test-go/testify/require"
)

func TestAssetScheduleDefaultTypeIsEffective(t *testing.T) {
	var schedules AssetsSchedule

	err := json.Unmarshal([]byte(`{"schedules":{"BTC":{"schedule":"0 0 * * 1-5","holidays":"[]"}}}`), &schedules)
	require.NoError(t, err)
	require.Equal(t, ScheduleTypeEffective, schedules.Schedules["BTC"].GetScheduleType())
	require.True(t, schedules.Schedules["BTC"].IsEffective())
}

func TestAssetScheduleInfoTypeParsing(t *testing.T) {
	var schedules AssetsSchedule

	err := json.Unmarshal([]byte(`{"schedules":{"SPCX":{"scheduleTimeZone":"America/New_York","schedule":"09:30-16:00|09:30-16:00|09:30-16:00|09:30-16:00|09:30-16:00|-|-","holidays":"[]","scheduleType":"info"}}}`), &schedules)
	require.NoError(t, err)
	require.Equal(t, ScheduleTypeInfo, schedules.Schedules["SPCX"].GetScheduleType())
	require.False(t, schedules.Schedules["SPCX"].IsEffective())
}

func TestAssetScheduleNilIsEffective(t *testing.T) {
	var schedule *AssetSchedule
	require.Equal(t, ScheduleTypeEffective, schedule.GetScheduleType())
	require.True(t, schedule.IsEffective())
}

func TestScheduleTypeIsEffective(t *testing.T) {
	require.True(t, ScheduleType("").IsEffective())
	require.True(t, ScheduleTypeEffective.IsEffective())
	require.False(t, ScheduleTypeInfo.IsEffective())
	require.False(t, ScheduleType("unknown").IsEffective())
}
