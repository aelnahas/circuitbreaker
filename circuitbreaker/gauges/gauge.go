package gauges

import (
	"errors"
	"fmt"
)

// Outcome is a type used to describe the different request outcomes
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

// Aggregate is used to keep track of the current performance of the outbound requests
type Aggregate struct {
	// Keep track of total requests in the snapshot
	RequestCount int
	// Keep track of requests that have failed
	FailureCount int
	// Keep track of request that succeeded
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

// Gauge provides an interface to logging and aggregating request results, mainly will be used
// by the state machine to be determine which state it is in
type Gauge interface {
	// Log outcom of a new request
	LogReading(Outcome)
	// Get Aggregate results so far
	OverallAggregate() Aggregate
	// Reset the gauge
	Reset()
}

var ErrEmptyMeasurements = errors.New("can read past record, no measurements taken")
