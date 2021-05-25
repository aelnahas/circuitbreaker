package circuitbreaker_test

import (
	"net/http"
	"testing"

	"github.com/aelnahas/circuitbreaker/circuitbreaker"
	"github.com/aelnahas/circuitbreaker/circuitbreaker/gauges"
)

type TestExecuter struct {
	ExecutorCalledCount int
	Handler             circuitbreaker.ExecuteHandler
}

func TestExecuteNilHandler(t *testing.T) {
	t.Run("NilExecuter", func(t *testing.T) {
		cb, err := circuitbreaker.NewBreaker("test")

		if err != nil {
			t.Errorf("circuitbreaker.NewMonitor, expected no err, got %s", err)
		}

		resp, err := cb.Execute(nil)
		expectedErr := circuitbreaker.ErrInvalidSettingParam{Param: "ExecuteHandler", Val: nil}

		if err != expectedErr {
			t.Errorf("Monitor.Execute(nil), error, expected %s, got %s", expectedErr, err)
		}

		if resp != nil {
			t.Errorf("Monitor.Execute(nil), resp, expected nil, got %+v", resp)
		}
	})
}

func TestExecuterNoRequestsPermitted(t *testing.T) {
	var IsSuccessful circuitbreaker.IsSuccessfulHandler = func(r *http.Response, e error) bool {
		return false
	}

	te := TestExecuter{}
	te.Handler = func(name string) (*http.Response, error) {
		te.ExecutorCalledCount++
		return &http.Response{}, nil
	}

	gauge := gauges.NewFixedWindowGauge(1)
	settings, _ := circuitbreaker.NewSettings("test", circuitbreaker.WithIsSuccessfulHandler(IsSuccessful), circuitbreaker.WithFailureRate(1), circuitbreaker.WithGauge(gauge))
	cb, err := circuitbreaker.NewBreakerWithSettings(settings)

	if err != nil {
		t.Errorf("Monitor.NewMonitor, expected no err, got %s", err)
	}

	cb.ForceState(circuitbreaker.Open)
	resp, err := cb.Execute(te.Handler)
	expectedErr := circuitbreaker.ErrRequestNotPermitted{Name: "test", State: circuitbreaker.Open}

	if err != expectedErr {
		t.Errorf("cb.Execute, error, expected : '%s', got : '%s'", expectedErr, err)

	}

	if resp != nil {
		t.Errorf("cb.Execute, resp, expected : nil, got : %+v", resp)
	}

	// should only be called once
	if te.ExecutorCalledCount > 0 {
		t.Errorf("cb.Execute, handler called N times, expected : 0, got: %d", te.ExecutorCalledCount)
	}
}

func TestExecuterRequestsPermitted(t *testing.T) {
	var IsSuccessful circuitbreaker.IsSuccessfulHandler = func(r *http.Response, e error) bool {
		return true
	}

	te := TestExecuter{}
	resp := &http.Response{}
	te.Handler = func(name string) (*http.Response, error) {
		te.ExecutorCalledCount++
		return resp, nil
	}

	gauge := gauges.NewFixedWindowGauge(1)
	settings, _ := circuitbreaker.NewSettings("test", circuitbreaker.WithIsSuccessfulHandler(IsSuccessful), circuitbreaker.WithFailureRate(1), circuitbreaker.WithGauge(gauge))
	cb, err := circuitbreaker.NewBreakerWithSettings(settings)

	if err != nil {
		t.Errorf("circuitbreaker.NewMonitor, expected no err, got %s", err)
	}

	actualResp, err := cb.Execute(te.Handler)

	if err != nil {
		t.Errorf("cb.Execute, error, expected : 'nil', got : '%s'", err)

	}

	if resp != actualResp {
		t.Errorf("cb.Execute, resp, expected : %+v, got : %+v", resp, actualResp)
	}

	// should only be called once
	if te.ExecutorCalledCount != 1 {
		t.Errorf("cb.Execute, handler called N times, expected : 0, got: %d", te.ExecutorCalledCount)
	}
}

func TestMonitorReset(t *testing.T) {
	var IsSuccessful circuitbreaker.IsSuccessfulHandler = func(r *http.Response, e error) bool {
		return true
	}

	te := TestExecuter{}
	resp := &http.Response{}
	te.Handler = func(name string) (*http.Response, error) {
		te.ExecutorCalledCount++
		return resp, nil
	}

	gauge := gauges.NewFixedWindowGauge(1)
	settings, _ := circuitbreaker.NewSettings("test", circuitbreaker.WithIsSuccessfulHandler(IsSuccessful), circuitbreaker.WithFailureRate(1), circuitbreaker.WithGauge(gauge))
	cb, err := circuitbreaker.NewBreakerWithSettings(settings)

	if err != nil {
		t.Errorf("Monitor.NewMonitor, expected no err, got %s", err)
	}

	cb.ForceState(circuitbreaker.Open)
	cb.Reset()

	actualResp, err := cb.Execute(te.Handler)

	if err != nil {
		t.Errorf("cb.Execute, error, expected : 'nil', got : '%s'", err)

	}

	if resp != actualResp {
		t.Errorf("cb.Execute, resp, expected : %+v, got : %+v", resp, actualResp)
	}

	// should only be called once
	if te.ExecutorCalledCount != 1 {
		t.Errorf("cb.Execute, handler called N times, expected : 0, got: %d", te.ExecutorCalledCount)
	}
}
