package circuitbreaker

import (
	"time"

	"github.com/aelnahas/circuitbreaker/circuitbreaker/gauges"
)

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

func (s State) String() string {
	switch s {
	case Closed:
		return "closed"
	case Open:
		return "open"
	case HalfOpen:
		return "half-open"
	default:
		return "unknown state"
	}
}

type stateMachine struct {
	state         State
	gauge         gauges.Gauge
	requestCount  int
	thresholds    Thresholds
	timer         *time.Timer
	onStateChange func(from, to State)
}

func NewStateMachine(gauge gauges.Gauge, thresholds Thresholds, onStateChange func(from, to State)) *stateMachine {
	return &stateMachine{
		gauge:         gauge,
		state:         Closed,
		thresholds:    thresholds,
		onStateChange: onStateChange,
	}
}

func (sm *stateMachine) ReportOutcome(outcome gauges.Outcome) (State, error) {
	if !sm.ShouldMakeRequests() {
		return sm.state, ErrRequestNotPermitted{State: sm.state}
	}

	sm.gauge.LogReading(outcome)
	if sm.state == HalfOpen && sm.requestCount <= sm.thresholds.MaxRequestOnHalfOpen {
		sm.requestCount++
	}
	sm.updateState()
	return sm.state, nil
}

func (sm *stateMachine) Reset() {
	sm.state = Closed
	sm.gauge.Reset()
}

func (sm *stateMachine) TransitionState(target State) {
	switch target {
	case Closed:
		sm.transitionToClosed()
	case HalfOpen:
		sm.transitionToHalfOpen()
	case Open:
		sm.transitionToOpen()
	}
}

func (sm *stateMachine) State() State {
	return sm.state
}

func (sm *stateMachine) ShouldMakeRequests() bool {
	switch sm.state {
	case Closed:
		return true
	case Open:
		return false
	case HalfOpen:
		return true
	default:
		return true
	}
}

func (sm *stateMachine) RequestCount() int {
	return sm.requestCount
}

func (sm *stateMachine) updateState() {
	metrics := sm.gauge.OverallAggregate()

	if metrics.RequestCount >= sm.thresholds.MinRequests && sm.state == Closed && metrics.FailureRate() > sm.thresholds.FailureRate {
		sm.TransitionState(Open)
	} else if sm.state == HalfOpen && sm.requestCount > sm.thresholds.MaxRequestOnHalfOpen {
		if metrics.SuccessRate() >= sm.thresholds.RecoveryRate {
			sm.TransitionState(Closed)
		} else {
			sm.TransitionState(Open)
		}
	}
}

func (sm *stateMachine) transitionToOpen() {
	sm.timer = time.AfterFunc(sm.thresholds.CooldownDuration, sm.watchCooldown)
	sm.state = Open
}

func (sm *stateMachine) transitionToHalfOpen() {
	sm.state = HalfOpen
	sm.gauge.Reset()
	sm.onStateChange(Open, HalfOpen)
}

func (sm *stateMachine) transitionToClosed() {
	sm.state = Closed
	sm.requestCount = 0
}

func (sm *stateMachine) watchCooldown() {
	sm.transitionToHalfOpen()
}
