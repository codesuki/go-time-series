package timeseries

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
)

// TODO: do table based testing

func TestRecent(t *testing.T) {
	clock := clock.NewMock()
	ts := &timeseries{clock: clock, levels: createLevels(clock, []time.Duration{time.Second, time.Minute})}

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
	ts := &timeseries{clock: clock, levels: createLevels(clock, []time.Duration{time.Second, time.Minute})}

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
	ts := &timeseries{clock: clock, levels: createLevels(clock, []time.Duration{time.Second, time.Minute})}

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
