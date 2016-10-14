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

	res, _ := ts.Recent(2 * time.Second)
	if res != 3 {
		t.Errorf("expected %d got %f", 3, res)
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

	res, _ := ts.Recent(2 * time.Minute)
	if res != 61 {
		t.Errorf("expected %d got %f", 61, res)
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

	// 60 + 1 + 60 * 1/60 (1 second of 1 minute bin) = 62
	res, _ := ts.Recent(60 * time.Minute)
	if res != 62 {
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

	// when querying the most recent time at a bigger granularity,
	// here 1m, there is some error because the count will be interpolated.

	// 60 + 1 + 60 * 1/60 (1 second of 1 minute bin) = 62
	res, _ := ts.Recent(120 * time.Minute)
	if res != 124 {
		t.Errorf("expected %d got %f", 62, res)
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
	clock.Add(time.Second)

	res, _ := ts.Recent(59 * time.Second)
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
	// (59/60 * 60) + (1/60 * 60) = 60
	res, _ := ts.Recent(time.Minute)
	if res != 60 {
		t.Errorf("expected %d got %f", 2, res)
	}
}

/*
func TestRecent(t *testing.T) {
	clock := clock.NewMock()
	ts := &timeseries{clock: clock, levels: createLevels(clock, []time.Duration{time.Second, time.Minute})}

	// time 12:00
	ts.Increase(2)
	clock.Add(time.Minute * 1) // time: 12:01:00
	ts.Increase(4)
	clock.Add(time.Minute * 1) // time: 12:02:00
	ts.Increase(6)
	clock.Add(time.Second * 1) // time: 12:02:01
	ts.Increase(1)
	clock.Add(time.Second * 10) // time: 12:02:11
	ts.Increase(2)
	clock.Add(time.Second * 10) // time: 12:02:21
	ts.Increase(2)
	clock.Add(time.Second * 10) // time: 12:02:31
	ts.Increase(2)
	clock.Add(time.Second * 10) // time: 12:02:41
	ts.Increase(2)
	clock.Add(time.Second * 10) // time: 12:02:51
	ts.Increase(2)
	clock.Add(time.Second * 9) // time: 12:03:00
	ts.Increase(9)
	clock.Add(time.Second * 1) // time: 12:03:01
	ts.Increase(1)

	fmt.Println(ts.Recent(time.Second * 30)) //==15
	fmt.Println(ts.Recent(time.Second * 1))  //==9
	fmt.Println(ts.Recent(time.Second * 60)) //==20
}

*/
/*
func TestIncrease2(t *testing.T) {
	clock := clock.NewMock()
	ts := &timeseries{clock: clock, levels: createLevels(clock, []time.Duration{time.Second, time.Minute})}

	// time 12:00
	ts.Increase(2)
	clock.Add(time.Minute * 60) // time: 13:00:00
	ts.Increase(3)

	fmt.Println(ts.Recent(time.Minute * 60)) //==2
}
*/
