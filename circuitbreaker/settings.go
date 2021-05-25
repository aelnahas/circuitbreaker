package circuitbreaker

import (
	"net/http"
	"time"

	"github.com/aelnahas/circuitbreaker/circuitbreaker/gauges"
)

//IsSUCcessfulHandler gets called back to determine if the response is a success
type IsSuccessfulHandler func(*http.Response, error) bool

//OnStateChangeHandler gets called back when circuit breaker switches states
type OnStateChangeHandler func(name string, from State, to State)

//Thresholds is a container to the different limits and important values that is used
//in the checking logic
type Thresholds struct {
	//FailureRate represents the ratio of failures to the request being made
	FailureRate float64
	//RecoveryRate determines if a circuit breaker should switch to closed from half-open depening on the success
	//of the request during half-open state
	RecoveryRate float64
	//CooldownDuration is the amount of time a circuit breaker will remain in the open state and not forward calls
	CooldownDuration time.Duration
	//MaxRequestOnHalfOpen Maximum number of requests permitted during the half-open state
	MaxRequestOnHalfOpen int
	//MinRequest minimum request before failing, this helps to combat initial spikes
	MinRequests int
}

//Settings is a collection of settings and options used with the circuit breaker and its internal members
type Settings struct {
	//Name is used to label the circuit breaker to help distinguish them
	Name string
	//Thresholds are as described above the limits of the circuit breaker, impacts the state transitions
	Thresholds Thresholds
	//IsSuccessful callback to help determin whether or not a request is successful
	IsSuccessful IsSuccessfulHandler
	//OnStateChange called back when states have transitioned
	OnStateChange OnStateChangeHandler
	//Gauge is used to collect metric to analyze the status of the requests
	Gauge gauges.Gauge
}

//DefaultFailureRate default failure rate set to 10%
const DefaultFailureRate float64 = 10.0

//DefaultRecoveryRate default recovery rate set to 10%
const DefaultRecoveryRate float64 = 10.0

//DefaultCooldownDuration default cool down duration set to 30 seconds
const DefaultCooldownDuration time.Duration = 30 * time.Second

//DefaultWindowSize used with FixedWindowGauge
const DefaultWindowSize int = 100

//DefaultMaxRequestOnHalfOpen default maximum request when circuit breaker is in half-open state
const DefaultMaxRequestOnHalfOpen int = 10

//DefaultMinRequest by default 10 request need to be executed before failing
const DefaultMinRequests int = 10

func DefaultIsSuccessful(resp *http.Response, err error) bool {
	return err == nil
}

//SettingsOption is a function that helps sets optional parameter of the settings
type SettingsOption func(*Settings)

func NewSettings(name string, opts ...SettingsOption) (*Settings, error) {
	thresholds := Thresholds{
		FailureRate:          DefaultFailureRate,
		RecoveryRate:         DefaultRecoveryRate,
		CooldownDuration:     DefaultCooldownDuration,
		MaxRequestOnHalfOpen: DefaultMaxRequestOnHalfOpen,
		MinRequests:          DefaultMinRequests,
	}
	settings := &Settings{
		Name:         name,
		Thresholds:   thresholds,
		IsSuccessful: DefaultIsSuccessful,
		Gauge:        gauges.NewFixedWindowGauge(DefaultWindowSize),
	}

	for _, opt := range opts {
		opt(settings)
	}

	if err := settings.Validate(); err != nil {
		return nil, err
	}

	return settings, nil
}

func (s *Settings) Validate() error {
	if s.Thresholds.CooldownDuration <= 0 {
		return ErrInvalidSettingParam{Param: "CooldownDuration", Val: s.Thresholds.CooldownDuration}
	}

	if s.Thresholds.FailureRate <= 0 || s.Thresholds.FailureRate > 100 {
		return ErrInvalidSettingParam{Param: "FailureRate", Val: s.Thresholds.FailureRate}
	}

	if s.Thresholds.RecoveryRate <= 0 || s.Thresholds.RecoveryRate > 100 {
		return ErrInvalidSettingParam{Param: "RecoveryRate", Val: s.Thresholds.RecoveryRate}
	}

	if s.Thresholds.MaxRequestOnHalfOpen <= 0 {
		return ErrInvalidSettingParam{Param: "MaxRequestOnHalfOpen", Val: s.Thresholds.MaxRequestOnHalfOpen}
	}

	if s.IsSuccessful == nil {
		return ErrInvalidSettingParam{Param: "IsSuccessful", Val: nil}
	}

	return nil
}

func WithFailureRate(rate float64) SettingsOption {
	return func(s *Settings) {
		s.Thresholds.FailureRate = rate
	}
}

func WithRecoveryRate(rate float64) SettingsOption {
	return func(s *Settings) {
		s.Thresholds.RecoveryRate = rate
	}
}

func WithCooldownDuration(duration time.Duration) SettingsOption {
	return func(s *Settings) {
		s.Thresholds.CooldownDuration = duration
	}
}

func WithIsSuccessfulHandler(handler IsSuccessfulHandler) SettingsOption {
	return func(s *Settings) {
		s.IsSuccessful = handler
	}
}

func WithOnStateChangeHandler(handler OnStateChangeHandler) SettingsOption {
	return func(s *Settings) {
		s.OnStateChange = handler
	}
}

func WithMaxRequestOnHalfOpen(max int) SettingsOption {
	return func(s *Settings) {
		s.Thresholds.MaxRequestOnHalfOpen = max
	}
}

func WithMinRequest(min int) SettingsOption {
	return func(s *Settings) {
		s.Thresholds.MinRequests = min
	}
}

func WithGauge(gauge gauges.Gauge) SettingsOption {
	return func(s *Settings) {
		s.Gauge = gauge
	}
}
