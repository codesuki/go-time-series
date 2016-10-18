package timeseries

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
)

// TODO: do table based testing

func setup() (*TimeSeries, *clock.Mock) {
	clock := clock.NewMock()
	ts, _ := NewTimeSeries(
		WithClock(clock),
		WithGranularities(
			[]Granularity{
				{time.Second, 60},
				{time.Minute, 60},
			},
		),
	)
	return ts, clock
}

func TestClock(t *testing.T) {
	clock := &defaultClock{}

	// there is a small chance this won't pass
	if clock.Now().Truncate(time.Second) != time.Now().Truncate(time.Second) {
		t.Errorf("default clock does not track time.Now")
	}
}

func TestNewTimeSeries(t *testing.T) {
	ts, err := NewTimeSeries()
	if ts == nil {
		t.Errorf("constructor returned nil")
	}
	if err != nil {
		t.Errorf("should not return error")
	}
}

func TestNewTimeSeriesWithGranularities(t *testing.T) {
	granularities := []Granularity{
		{time.Second, 60},
		{time.Minute, 60},
		{time.Hour, 24},
	}
	ts, err := NewTimeSeries(WithGranularities(granularities))
	if ts == nil || err != nil {
		t.Error("could not create time series")
	}

	badGranularities := []Granularity{
		{time.Minute, 60},
		{time.Second, 60},
		{time.Hour, 24},
	}
	_, err = NewTimeSeries(WithGranularities(badGranularities))
	if err != ErrBadGranularities {
		t.Error("should not accept decreasing granularities")
	}

	badGranularities = []Granularity{
		{time.Minute, 60},
		{time.Second, 0},
		{time.Hour, 24},
	}
	_, err = NewTimeSeries(WithGranularities(badGranularities))
	if err != ErrBadGranularities {
		t.Error("should not accept granularities with zero count")
	}

	_, err = NewTimeSeries(WithGranularities([]Granularity{}))
	if err != ErrBadGranularities {
		t.Error("should not accept empty granularities")
	}
}

func TestNewTimeSeriesWithClock(t *testing.T) {
	clock := clock.NewMock()
	ts, _ := NewTimeSeries(WithClock(clock))

	ts.Increase(2)
	clock.Add(time.Second * 1)
	ts.Increase(1)

	res, _ := ts.Range(time.Unix(0, 0), time.Unix(1, 0))
	if res != 2 {
		t.Errorf("expected %d got %f", 2, res)
	}
}

func TestRecentSeconds(t *testing.T) {
	ts, clock := setup()

	clock.Add(time.Minute * 5)
	ts.Increase(1)
	clock.Add(time.Second * 1)
	ts.Increase(2)
	clock.Add(time.Second * 1)
	ts.Increase(3)

	res, _ := ts.Recent(time.Second)
	if res != 2 {
		t.Errorf("expected %d got %f", 2, res)
	}

	res, _ = ts.Recent(2 * time.Second)
	if res != 3 {
		t.Errorf("expected %d got %f", 3, res)
	}

	// test earliest second
	clock.Add(57 * time.Second) // time: 09:05:59
	res, _ = ts.Recent(59 * time.Second)
	if res != 6 {
		t.Errorf("expected %d got %f", 6, res)
	}

	// test future time
	clock.Add(1 * time.Second)
	clock.Add(57 * time.Second) // time: 09:06:00
	res, _ = ts.Recent(59 * time.Second)
	if res != 0 {
		t.Errorf("expected %d got %f", 0, res)
	}
}

func TestRecentMinutes(t *testing.T) {
	ts, clock := setup()

	clock.Add(time.Minute * 1) // 09:01:00
	ts.Increase(60)
	clock.Add(time.Minute * 1) // 09:02:00
	ts.Increase(1)
	clock.Add(time.Minute * 1) // 09:03:00
	ts.Increase(60)
	clock.Add(time.Second * 1) // 09:03:01
	ts.Increase(3)

	// test interpolation at beginning
	// 59/60 * 60 + 1 + 60 = 120
	res, _ := ts.Recent(2 * time.Minute)
	if res != 120 {
		t.Errorf("expected %d got %f", 120, res)
	}

	// test interpolation at end
	// 60/2 = 30
	res, _ = ts.Range(
		clock.Now().Add(-2*time.Minute+-1*time.Second),  // 09:01:00
		clock.Now().Add(-1*time.Minute+-31*time.Second), // 09:01:30
	)
	if res != 30 {
		t.Errorf("expected %d got %f", 30, res)
	}

	// get from earliest data point
	clock.Add(time.Second*59 + time.Minute*56)
	ts.Increase(60)
	clock.Add(time.Minute * 1)
	ts.Increase(70)
	clock.Add(time.Minute * 59)
	res, _ = ts.Recent(time.Minute * 60)
	if res != 70 {
		t.Errorf("expected %d got %f", 70, res)
	}
}

func TestRecentWholeRange(t *testing.T) {
	ts, clock := setup()

	clock.Add(time.Minute * 1) // 09:01:00
	ts.Increase(60)
	clock.Add(time.Minute * 1) // 09:02:00
	ts.Increase(1)
	clock.Add(time.Minute * 1) // 09:03:00
	ts.Increase(60)
	clock.Add(time.Second * 1) // 09:03:01
	ts.Increase(3)

	// 60 + 1 + 60 = 121
	res, _ := ts.Recent(60 * time.Minute)
	if res != 121 {
		t.Errorf("expected %d got %f", 62, res)
	}
}

func TestRecentWholeRangeBig(t *testing.T) {
	ts, clock := setup()

	clock.Add(time.Minute * 1) // 09:01:00
	ts.Increase(60)
	clock.Add(time.Minute * 1) // 09:02:00
	ts.Increase(1)
	clock.Add(time.Minute * 1) // 09:03:00
	ts.Increase(60)
	clock.Add(time.Second * 1) // 09:03:01
	ts.Increase(3)

	// 60 + 1 + 60 = 121
	res, _ := ts.Recent(120 * time.Minute)
	if res != 121 {
		t.Errorf("expected %d got %f", 121, res)
	}
}

func TestRangeEndInFuture(t *testing.T) {
	ts, clock := setup()

	clock.Add(time.Minute * 1) // 09:01:00
	ts.Increase(1)

	res, _ := ts.Range(clock.Now().Add(-1*time.Minute), clock.Now().Add(5*time.Minute))
	if res != 0 {
		t.Errorf("expected %d got %f", 0, res)
	}
}

func TestRangeBadRange(t *testing.T) {
	ts, clock := setup()

	clock.Add(time.Minute * 1) // 09:01:00
	ts.Increase(60)
	clock.Add(time.Minute * 1) // 09:02:00
	ts.Increase(1)
	clock.Add(time.Minute * 1) // 09:03:00
	ts.Increase(60)
	clock.Add(time.Second * 1) // 09:03:01
	ts.Increase(3)

	// start is after end
	_, err := ts.Range(clock.Now().Add(time.Minute), clock.Now())
	if err != ErrBadRange {
		t.Errorf("should return ErrBadRange")
	}

	// range is after end
	_, err = ts.Range(clock.Now().Add(time.Minute), clock.Now().Add(5*time.Minute))
	if err != ErrRangeNotCovered {
		t.Errorf("should return ErrRangeNotCovered")
	}

	// range is before start
	_, err = ts.Range(clock.Now().Add(-5*time.Hour), clock.Now().Add(-4*time.Hour))
	if err != ErrRangeNotCovered {
		t.Errorf("should return ErrRangeNotCovered")
	}
}

func TestIncrease(t *testing.T) {
	ts, clock := setup()

	// time 12:00
	ts.Increase(2)
	clock.Add(time.Minute * 1) // time: 12:01:00
	ts.Increase(4)
	clock.Add(time.Minute * 1) // time: 12:02:00
	ts.Increase(6)
	clock.Add(time.Second * 10) // time: 12:02:10
	ts.Increase(2)
	clock.Add(time.Second * 10) // time: 12:02:20
	ts.Increase(2)
	clock.Add(time.Second * 10) // time: 12:02:30
	ts.Increase(2)
	clock.Add(time.Second * 10) // time: 12:02:40
	ts.Increase(2)
	clock.Add(time.Second * 10) // time: 12:02:50
	ts.Increase(2)
	clock.Add(time.Second * 10) // time: 12:03:00
	ts.Increase(2)
	// get range from 12:00:30 - 12:02:30
	// 0.5 * 2 + 4 + 0.5 * 16 = 13
	res, _ := ts.Range(clock.Now().Add(-time.Second*150), clock.Now().Add(-time.Second*30))
	if res != 13 {
		t.Errorf("expected %d got %f", 13, res)
	}

	// get range from 12:01:00 - 12:02:00
	// = 4
	res, _ = ts.Range(clock.Now().Add(-time.Second*120), clock.Now().Add(-time.Second*60))
	if res != 4 {
		t.Errorf("expected %d got %f", 4, res)
	}

}

func TestIncreasePending(t *testing.T) {
	ts, clock := setup()

	ts.Increase(1) // this should advance and reset pending
	ts.Increase(1) // this should increase pending
	clock.Add(time.Second)
	ts.Increase(1)

	res, _ := ts.Recent(59 * time.Second)
	if res != 2 {
		t.Errorf("expected %d got %f", 2, res)
	}

	clock.Add(time.Second) // the latest data gets merged in because time advanced

	res, _ = ts.Recent(59 * time.Second)
	if res != 3 {
		t.Errorf("expected %d got %f", 3, res)
	}
}

func TestIncreaseAtTime(t *testing.T) {
	ts, clock := setup()

	ts.Increase(60)                                        // time: 09:00:00
	clock.Add(time.Second)                                 // time: 09:00:01
	ts.IncreaseAtTime(60, clock.Now().Add(-1*time.Minute)) // time: 08:59:01
	ts.Increase(1)                                         // time: 09:00:01

	// from: 08:59:01 - 09:00:01
	// (59/60 * 60) + 60 = 119
	res, _ := ts.Recent(time.Minute)
	if res != 119 {
		t.Errorf("expected %d got %f", 119, res)
	}

	// from: 08:59:00 - 09:00:00
	// 60
	res, _ = ts.Range(
		clock.Now().Add(-1*time.Minute+-1*time.Second),
		clock.Now().Add(-1*time.Second),
	)
	if res != 60 {
		t.Errorf("expected %d got %f", 60, res)
	}
}
