package circuitbreaker

import (
	"net/http"
	"sync"
)

type ExecuteHandler func(name string) (*http.Response, error)

type RequestInterceptor struct {
	mutex        sync.RWMutex
	Settings     *Settings
	stateMachine *stateMachine
}

func NewRequestInterceptor(name string) (*RequestInterceptor, error) {
	settings, err := NewSettings(name)
	if err != nil {
		return nil, err
	}

	return NewRequestInterceptorWithSettings(settings)
}

func NewRequestInterceptorWithSettings(settings *Settings) (*RequestInterceptor, error) {
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	ri := &RequestInterceptor{
		Settings:     settings,
		stateMachine: NewStateMachine(settings.WindowSize, settings.Thresholds),
	}
	return ri, nil
}

func (ri *RequestInterceptor) Execute(handler ExecuteHandler) (*http.Response, error) {
	ri.mutex.Lock()
	defer ri.mutex.Unlock()

	if handler == nil {
		return nil, ErrInvalidSettingParam{Param: "ExecuteHandler", Val: nil}
	}

	prev := ri.stateMachine.State()
	if !ri.stateMachine.ShouldMakeRequests() {
		return nil, ErrRequestNotPermitted{
			Name:  ri.Settings.Name,
			State: ri.stateMachine.State(),
		}
	}

	resp, err := handler(ri.Settings.Name)

	var outcome Outcome

	if ri.Settings.IsSuccessful(resp, err) {
		outcome = Succeeded
	} else {
		outcome = Failed
	}

	state, err := ri.stateMachine.ReportOutcome(outcome)

	if err != nil {
		return nil, err
	}

	if state != prev {
		ri.Settings.OnStateChange(ri.Settings.Name, prev, state, ri.stateMachine.metricRecorder.Metrics())
	}

	return resp, err
}

func (ri *RequestInterceptor) Reset() {
	ri.mutex.Lock()
	defer ri.mutex.Unlock()
	ri.stateMachine.Reset()
}

func (ri *RequestInterceptor) ForceState(state State) {
	ri.mutex.Lock()
	defer ri.mutex.Unlock()
	ri.stateMachine.TransitionState(state)
}
