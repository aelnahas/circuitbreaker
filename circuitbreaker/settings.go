package circuitbreaker

import (
	"net/http"
	"time"
)

type IsSuccessfulHandler func(*http.Response, error) bool
type CallbackHandler func(*http.Response, error) (interface{}, error)
type OnStateChangeHandler func(name string, from State, to State, metrics Metrics)

type Thresholds struct {
	FailureRate          float64
	RecoveryRate         float64
	CooldownDuration     time.Duration
	MaxRequestOnHalfOpen int
}

type Settings struct {
	Name          string
	WindowSize    int
	Thresholds    Thresholds
	IsSuccessful  IsSuccessfulHandler
	OnStateChange OnStateChangeHandler
}

const DefaultFailureRate float64 = 10.0
const DefaultRecoveryRate float64 = 10.0
const DefaultCooldownDuration time.Duration = 30 * time.Second
const DefaultWindowSize int = 100
const DefaultMaxRequestOnHalfOpen int = 10

func DefaultIsSuccessful(resp *http.Response, err error) bool {
	return err == nil
}

type SettingsOption func(*Settings)

func NewSettings(name string, opts ...SettingsOption) (*Settings, error) {
	thresholds := Thresholds{
		FailureRate:          DefaultFailureRate,
		RecoveryRate:         DefaultRecoveryRate,
		CooldownDuration:     DefaultCooldownDuration,
		MaxRequestOnHalfOpen: DefaultMaxRequestOnHalfOpen,
	}
	settings := &Settings{
		Name:         name,
		Thresholds:   thresholds,
		WindowSize:   DefaultWindowSize,
		IsSuccessful: DefaultIsSuccessful,
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

	if s.WindowSize <= 0 {
		return ErrInvalidSettingParam{Param: "WindowSize", Val: s.WindowSize}
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

func WithWindowSize(size int) SettingsOption {
	return func(s *Settings) {
		s.WindowSize = size
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
