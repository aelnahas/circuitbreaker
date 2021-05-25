package gauges

import (
	"errors"
	"fmt"
)

type Outcome int

const (
	Success Outcome = iota + 1
	Failure
)

func (o Outcome) String() string {
	switch o {
	case Success:
		return "success"
	case Failure:
		return "failure"
	default:
		return "unknown"
	}
}

type Aggregate struct {
	RequestCount int
	FailureCount int
	SuccessCount int
}

func (a *Aggregate) record(outcome Outcome) {
	a.RequestCount++

	if outcome == Success {
		a.SuccessCount++
	} else {
		a.FailureCount++
	}
}

func (a *Aggregate) erase(reading *Aggregate) {
	a.RequestCount -= reading.RequestCount
	a.FailureCount -= reading.FailureCount
	a.SuccessCount -= reading.SuccessCount
}

func (a *Aggregate) reset() {
	a.FailureCount = 0
	a.SuccessCount = 0
	a.RequestCount = 0
}

func (a *Aggregate) FailureRate() float64 {
	fmt.Println(a.FailureCount)
	if a.RequestCount > 0 {
		return 100 * float64(a.FailureCount) / float64(a.RequestCount)
	}
	return 0
}

func (a *Aggregate) SuccessRate() float64 {
	if a.RequestCount > 0 {
		return 100 * float64(a.SuccessCount) / float64(a.RequestCount)
	}
	return 0
}

type Gauge interface {
	LogReading(Outcome)
	OverallAggregate() Aggregate
	Reset()
}

var ErrEmptyMeasurements = errors.New("can read past record, no measurements taken")
