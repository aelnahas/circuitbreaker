package gauges

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

type partialAggregate struct {
	timestamp time.Time
	expiresAt time.Time
	aggregate *Aggregate
}

func newPartial(timestamp time.Time) *partialAggregate {
	return &partialAggregate{
		timestamp: timestamp,
		expiresAt: timestamp.Add(30 * time.Second),
		aggregate: &Aggregate{},
	}
}

type TimeBasedGauge struct {
	lock      sync.RWMutex
	timeRange time.Duration

	partialAggregates *list.List
	totalAggregate    *Aggregate
}

var _ Gauge = &TimeBasedGauge{}

func NewTimeBasedGauge(timeRange time.Duration) *TimeBasedGauge {
	gauge := &TimeBasedGauge{
		timeRange: timeRange,

		partialAggregates: list.New(),
		totalAggregate:    &Aggregate{},
	}

	go gauge.runBackgroundTicker()
	return gauge
}

func (g *TimeBasedGauge) Reset() {
	g.totalAggregate = &Aggregate{}
	g.partialAggregates.Init()
}

func (g *TimeBasedGauge) LogReading(reading Outcome) {

	partial := g.fastForwardToCurrentPartial()
	g.totalAggregate.record(reading)
	partial.aggregate.record(reading)

	fmt.Println(g.totalAggregate)
}

func (g *TimeBasedGauge) OverallAggregate() Aggregate {
	g.fastForwardToCurrentPartial()
	return *g.totalAggregate
}

func (g *TimeBasedGauge) fastForwardToCurrentPartial() *partialAggregate {
	g.lock.Lock()
	defer g.lock.Unlock()
	currentTime := time.Now()

	if g.partialAggregates.Len() == 0 {
		partial := newPartial(currentTime)
		g.partialAggregates.PushBack(partial)
		return partial
	}

	latest := g.partialAggregates.Front()
	latestPartial := latest.Value.(*partialAggregate)

	if latestPartial.expiresAt.After(currentTime) {
		partial := newPartial(currentTime)
		g.partialAggregates.PushBack(partial)
		return partial
	}

	return latestPartial
}

func (g *TimeBasedGauge) removeExpiredPartial() {
	g.lock.Lock()
	defer g.lock.Unlock()
	oldest := g.partialAggregates.Front()
	partial := oldest.Value.(*partialAggregate)

	if partial.expiresAt.Before(time.Now()) {
		g.partialAggregates.Remove(oldest)
		g.totalAggregate.erase(partial.aggregate)
	}
}

func (g *TimeBasedGauge) runBackgroundTicker() {
	ticker := time.NewTicker(2 * time.Second)

	for _ = range ticker.C {
		if g.partialAggregates.Len() > 0 {
			g.removeExpiredPartial()
		}
	}
}
