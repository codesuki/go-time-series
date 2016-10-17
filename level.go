package timeseries

import (
	"log"
	"time"
)

type level struct {
	clock       Clock
	granularity time.Duration
	length      int
	end         time.Time
	oldest      int
	newest      int
	buckets     []int
}

func newLevel(clock Clock, granularity time.Duration, length int) level {
	level := level{clock: clock, granularity: granularity, length: length}
	level.init()
	return level
}

func (l *level) init() {
	buckets := make([]int, l.length)
	l.buckets = buckets
	l.clear(time.Time{})
}

func (l *level) clear(time time.Time) {
	l.oldest = 1
	l.newest = 0
	l.end = time.Truncate(l.granularity)
	for i := range l.buckets {
		l.buckets[i] = 0
	}
}

func (l *level) duration() time.Duration {
	return l.granularity*time.Duration(l.length) - l.granularity
}

func (l *level) earliest() time.Time {
	return l.end.Add(-l.duration())
}

func (l *level) latest() time.Time {
	return l.end
}

func (l *level) increaseAtTime(amount int, time time.Time) {
	difference := l.end.Sub(time.Truncate(l.granularity))
	if difference < 0 {
		// this cannot be negative because we advance before
		// can at least be 0
		log.Println("level.increaseTime was called with a time in the future")
	}
	// l.length-1 because the newest element is always l.length-1 away from oldest
	steps := (l.length - 1) - int(difference/l.granularity)
	index := (l.oldest + steps) % l.length
	l.buckets[index] += amount
}

func (l *level) advance(target time.Time) {
	if !l.end.Before(target) {
		return
	}
	for target.After(l.end) {
		l.end = l.end.Add(l.granularity)
		l.buckets[l.oldest] = 0
		l.newest = l.oldest
		l.oldest = (l.oldest + 1) % len(l.buckets)
	}
}

// TODO: find a better way to handle latest parameter
// The parameter is used to avoid the overlap computation if end overlaps with the current time.
// Probably will find away when implementing redis version.
func (l *level) sumInterval(start, end time.Time, latest time.Time) float64 {
	if start.Before(l.earliest()) {
		start = l.earliest()
	}
	if end.After(l.latest()) {
		end = l.latest()
	}
	idx := 0
	// this is how many time steps start is away from earliest
	startSteps := start.Sub(l.earliest()) / l.granularity
	idx += int(startSteps)

	currentTime := l.earliest()
	currentTime = currentTime.Add(startSteps * l.granularity)

	sum := 0.0
	for idx < l.length && currentTime.Before(end) {
		nextTime := currentTime.Add(l.granularity)
		if nextTime.After(latest) {
			nextTime = latest
		}
		if nextTime.Before(start) {
			// the case nextTime.Before(start) happens when start is after latest
			// therefore we don't have data and can return
			break
		}
		count := float64(l.buckets[(l.oldest+idx)%l.length])
		if currentTime.Before(start) || nextTime.After(end) {
			// current bucket overlaps time range
			overlapStart := max(currentTime, start)
			overlapEnd := min(nextTime, end)
			overlap := overlapEnd.Sub(overlapStart).Seconds() / l.granularity.Seconds()
			count *= overlap
		}
		sum += count
		idx++
		currentTime = currentTime.Add(l.granularity)
	}
	return sum
}

func min(t1, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t1
	}
	return t2
}

func max(t1, t2 time.Time) time.Time {
	if t1.After(t2) {
		return t1
	}
	return t2
}
