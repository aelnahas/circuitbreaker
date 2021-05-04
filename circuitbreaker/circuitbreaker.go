package circuitbreaker

import (
	"net/http"
	"sync"
)

type ExecuteHandler func(name string) (*http.Response, error)

type CircuitBreaker struct {
	mutex        sync.RWMutex
	Settings     *Settings
	stateMachine *stateMachine
}

func NewCircuitBreaker(name string) (*CircuitBreaker, error) {
	settings, err := NewSettings(name)
	if err != nil {
		return nil, err
	}

	return NewCircuitBreakerWithSettings(settings)
}

func NewCircuitBreakerWithSettings(settings *Settings) (*CircuitBreaker, error) {
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	cb := &CircuitBreaker{
		Settings:     settings,
		stateMachine: NewStateMachine(settings.WindowSize, settings.Thresholds),
	}
	return cb, nil
}

func (cb *CircuitBreaker) Execute(handler ExecuteHandler) (*http.Response, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if handler == nil {
		return nil, ErrInvalidSettingParam{Param: "ExecuteHandler", Val: nil}
	}

	prev := cb.stateMachine.State()
	if !cb.stateMachine.ShouldMakeRequests() {
		return nil, ErrRequestNotPermitted{
			Name:  cb.Settings.Name,
			State: cb.stateMachine.State(),
		}
	}

	resp, err := handler(cb.Settings.Name)

	var outcome Outcome

	if cb.Settings.IsSuccessful(resp, err) {
		outcome = Succeeded
	} else {
		outcome = Failed
	}

	state, err := cb.stateMachine.ReportOutcome(outcome)

	if err != nil {
		return nil, err
	}

	if state != prev {
		cb.Settings.OnStateChange(cb.Settings.Name, prev, state, cb.stateMachine.metricRecorder.Metrics())
	}

	return resp, err
}

func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.stateMachine.Reset()
}

func (cb *CircuitBreaker) ForceState(state State) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.stateMachine.TransitionState(state)
}
