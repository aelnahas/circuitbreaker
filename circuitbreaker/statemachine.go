package circuitbreaker

import (
	"time"
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
	state          State
	metricRecorder MetricsRecorder
	requestCount   int
	thresholds     Thresholds
	timer          *time.Timer
}

func NewStateMachine(windowSize int, thresholds Thresholds) *stateMachine {
	return &stateMachine{
		metricRecorder: NewFixedSizeSlidingWindowMetrics(windowSize),
		state:          Closed,
		thresholds:     thresholds,
	}
}

func NewMachineWithMetricRecorder(recorder MetricsRecorder, thresholds Thresholds) *stateMachine {
	return &stateMachine{
		metricRecorder: recorder,
		state:          Closed,
		thresholds:     thresholds,
	}
}

func (sm *stateMachine) ReportOutcome(outcome Outcome) (State, error) {
	if !sm.ShouldMakeRequests() {
		return sm.state, ErrRequestNotPermitted{State: sm.state}
	}
	sm.metricRecorder.Record(outcome)
	if sm.state == HalfOpen && sm.requestCount <= sm.thresholds.MaxRequestOnHalfOpen {
		sm.requestCount++
	}
	sm.updateState()
	return sm.state, nil
}

func (sm *stateMachine) Reset() {
	sm.state = Closed
	sm.metricRecorder.Reset()
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
	metrics := sm.metricRecorder.Metrics()

	if sm.state == Closed && metrics.FailureRate > sm.thresholds.FailureRate {
		sm.TransitionState(Open)
	} else if sm.state == HalfOpen && sm.requestCount > sm.thresholds.MaxRequestOnHalfOpen {
		if metrics.SuccessRate >= sm.thresholds.RecoveryRate {
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
	sm.metricRecorder.Reset()
}

func (sm *stateMachine) transitionToClosed() {
	sm.state = Closed
	sm.requestCount = 0
}

func (sm *stateMachine) watchCooldown() {
	sm.transitionToHalfOpen()
}
