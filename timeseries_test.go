package timeseries

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
)

// TODO: do table based testing

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
	granularities := []time.Duration{
		time.Second,
		time.Minute,
		time.Hour,
	}
	ts, err := NewTimeSeriesWithGranularities(granularities)
	if ts == nil || err != nil {
		t.Error("could not create time series")
	}

	badGranularities := []time.Duration{
		time.Minute,
		time.Second,
		time.Hour,
	}
	_, err = NewTimeSeriesWithGranularities(badGranularities)
	if err != ErrBadGranularities {
		t.Error("should not accept decreasing granularities")
	}

	_, err = NewTimeSeriesWithGranularities([]time.Duration{})
	if err != ErrBadGranularities {
		t.Error("should not accept empty granularities")
	}
}

func TestNewTimeSeriesWithClock(t *testing.T) {
	clock := clock.NewMock()
	ts, _ := NewTimeSeriesWithClock(clock)

	ts.Increase(2)
	clock.Add(time.Second * 1)
	ts.Increase(1)

	res, _ := ts.Range(time.Unix(0, 0), time.Unix(1, 0))
	if res != 2 {
		t.Errorf("expected %d got %f", 2, res)
	}
}

func TestRecent(t *testing.T) {
	clock := clock.NewMock()
	ts, _ := NewTimeSeriesWithGranularitiesAndClock([]time.Duration{time.Second, time.Minute}, clock)

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

func TestRecent2(t *testing.T) {
	clock := clock.NewMock()
	ts, _ := NewTimeSeriesWithGranularitiesAndClock([]time.Duration{time.Second, time.Minute}, clock)

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

func TestIncrease(t *testing.T) {
	clock := clock.NewMock()
	ts, _ := NewTimeSeriesWithGranularitiesAndClock([]time.Duration{time.Second, time.Minute}, clock)

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
	clock := clock.NewMock()
	ts, _ := NewTimeSeriesWithGranularitiesAndClock([]time.Duration{time.Second, time.Minute}, clock)

	ts.Increase(1) // this should advance and reset pending
	ts.Increase(1) // this should increase pending
	clock.Add(time.Second)
	ts.Increase(1)

	res, _ := ts.Recent(time.Minute)
	if res != 2 {
		t.Errorf("expected %d got %f", 2, res)
	}
}

func TestIncreaseAtTime(t *testing.T) {
	clock := clock.NewMock()
	ts, _ := NewTimeSeriesWithGranularitiesAndClock([]time.Duration{time.Second, time.Minute}, clock)

	ts.Increase(1)
	clock.Add(time.Second)
	ts.IncreaseAtTime(1, clock.Now().Add(-1*time.Minute))
	ts.Increase(1)

	res, _ := ts.Recent(time.Minute)
	if res != 2 {
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
