package circuitbreaker

import (
	"net/http"
	"sync"

	"github.com/aelnahas/circuitbreaker/circuitbreaker/gauges"
)

// ExecuteHandler describe a function that handles triggering the actual requests
type ExecuteHandler func(name string) (*http.Response, error)

// Breaker manages the circuit breaker activities such as executing the request
type Breaker struct {
	mutex        sync.RWMutex
	Settings     *Settings
	stateMachine *stateMachine
}

func NewBreaker(name string) (*Breaker, error) {
	settings, err := NewSettings(name)
	if err != nil {
		return nil, err
	}

	return NewBreakerWithSettings(settings)
}

func NewBreakerWithSettings(settings *Settings) (*Breaker, error) {
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	b := &Breaker{
		Settings: settings,
	}
	b.stateMachine = NewStateMachine(settings.Gauge, settings.Thresholds, b.onStateChange)
	return b, nil
}

func (b *Breaker) Execute(handler ExecuteHandler) (*http.Response, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if handler == nil {
		return nil, ErrInvalidSettingParam{Param: "ExecuteHandler", Val: nil}
	}

	prev := b.stateMachine.State()
	if !b.stateMachine.ShouldMakeRequests() {
		return nil, ErrRequestNotPermitted{
			Name:  b.Settings.Name,
			State: b.stateMachine.State(),
		}
	}

	resp, err := handler(b.Settings.Name)

	var outcome gauges.Outcome

	if b.Settings.IsSuccessful(resp, err) {
		outcome = gauges.Success
	} else {
		outcome = gauges.Failure
	}

	state, err := b.stateMachine.ReportOutcome(outcome)

	if state != prev {
		b.Settings.OnStateChange(b.Settings.Name, prev, state)
	}

	return resp, err
}

func (b *Breaker) Reset() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.stateMachine.Reset()
}

func (b *Breaker) ForceState(state State) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.stateMachine.TransitionState(state)
}

func (b *Breaker) onStateChange(from, to State) {
	b.Settings.OnStateChange(b.Settings.Name, from, to)
}
