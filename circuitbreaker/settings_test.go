package circuitbreaker_test

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/aelnahas/circuitbreaker/circuitbreaker"
)

func defaultSettings(name string) *circuitbreaker.Settings {
	return &circuitbreaker.Settings{
		Name: name,
		Thresholds: circuitbreaker.Thresholds{
			MaxRequestOnHalfOpen: circuitbreaker.DefaultMaxRequestOnHalfOpen,
			FailureRate:          circuitbreaker.DefaultFailureRate,
			RecoveryRate:         circuitbreaker.DefaultRecoveryRate,
			CooldownDuration:     circuitbreaker.DefaultCooldownDuration,
		},
		IsSuccessful: circuitbreaker.DefaultIsSuccessful,
	}
}

func compareSettings(t *testing.T, expected, actual *circuitbreaker.Settings) {
	t.Run("TestThresholds", func(t *testing.T) {
		t.Run("TestCooldownDuration", func(t *testing.T) {
			if expected.Thresholds.CooldownDuration != actual.Thresholds.CooldownDuration {
				t.Errorf("expected %d, got %d", expected.Thresholds.CooldownDuration, actual.Thresholds.CooldownDuration)
			}
		})

		t.Run("TestFailureRate", func(t *testing.T) {
			if expected.Thresholds.FailureRate != actual.Thresholds.FailureRate {
				t.Errorf("expected %f, got %f", expected.Thresholds.FailureRate, actual.Thresholds.FailureRate)
			}
		})

		t.Run("TestRecoveryRate", func(t *testing.T) {
			if expected.Thresholds.RecoveryRate != actual.Thresholds.RecoveryRate {
				t.Errorf("expected %f, got %f", expected.Thresholds.RecoveryRate, actual.Thresholds.RecoveryRate)
			}
		})

		t.Run("TestMaxRequestOnHalfOpen", func(t *testing.T) {
			if expected.Thresholds.MaxRequestOnHalfOpen != actual.Thresholds.MaxRequestOnHalfOpen {
				t.Errorf("expected %d, got %d", expected.Thresholds.MaxRequestOnHalfOpen, actual.Thresholds.MaxRequestOnHalfOpen)
			}
		})
	})

	t.Run("TestName", func(t *testing.T) {
		if expected.Name != actual.Name {
			t.Errorf("Settings.Name expected %s, got %s", expected.Name, actual.Name)
		}
	})

	t.Run("TestIsSuccessful", func(t *testing.T) {
		var expectedPtr uintptr = reflect.ValueOf(expected.IsSuccessful).Pointer()
		var actualPtr uintptr = reflect.ValueOf(actual.IsSuccessful).Pointer()

		if expectedPtr != actualPtr {
			t.Error("Settings.IsSuccessful expected and actual settings have different handlers")
		}

	})
}

func TestDefaultSettings(t *testing.T) {
	name := "default"

	actual, err := circuitbreaker.NewSettings(name)

	expected := defaultSettings(name)

	if err != nil {
		t.Errorf("NewSetttings(%s) expected no errors got %s", name, err.Error())
	}

	compareSettings(t, expected, actual)
}

func TestWithFailureRate(t *testing.T) {
	t.Run("Test_WithValidFailureRate", func(t *testing.T) {
		name := "withFailureRate"
		expected := defaultSettings(name)
		failureRate := 70.0
		expected.Thresholds.FailureRate = failureRate

		settings, err := circuitbreaker.NewSettings(name, circuitbreaker.WithFailureRate(failureRate))

		if err != nil {
			t.Errorf("NewSetttings(%s, circuitbreaker.FailureRate(100)) expected no errors got %s", name, err.Error())
		}

		compareSettings(t, expected, settings)
	})

	t.Run("Test_InvalidRate", func(t *testing.T) {
		invalidValues := []float64{-1, 0}

		for _, val := range invalidValues {
			settings, err := circuitbreaker.NewSettings("test", circuitbreaker.WithFailureRate(val))

			expectedError := circuitbreaker.ErrInvalidSettingParam{Param: "FailureRate", Val: val}

			if err != expectedError {
				t.Fatalf("Test_InvalidRate, rate = %f, expected error to be '%s', got '%s'", val, expectedError, err)
			}
			if settings != nil {
				t.Fatalf("Test_InvalidRate, rate = %f, expected function to return nil settings, got %+v", val, settings)
			}
		}

	})
}

func TestWithRecoveryRate(t *testing.T) {
	t.Run("Test_WithValidRecoveryRate", func(t *testing.T) {
		name := "withRecoveryRate"
		expected := defaultSettings(name)
		rate := 70.0
		expected.Thresholds.RecoveryRate = rate

		settings, err := circuitbreaker.NewSettings(name, circuitbreaker.WithRecoveryRate(rate))

		if err != nil {
			t.Errorf("NewSetttings(%s, circuitbreaker.RecoveryRate(%f)) expected no errors got %s", name, rate, err.Error())
		}

		compareSettings(t, expected, settings)
	})

	t.Run("Test_InvalidRate", func(t *testing.T) {
		invalidValues := []float64{-1, 0}

		for _, val := range invalidValues {
			settings, err := circuitbreaker.NewSettings("test", circuitbreaker.WithRecoveryRate(val))

			expectedError := circuitbreaker.ErrInvalidSettingParam{Param: "RecoveryRate", Val: val}

			if err != expectedError {
				t.Fatalf("Test_InvalidRate, rate = %f, expected error to be '%s', got '%s'", val, expectedError, err)
			}
			if settings != nil {
				t.Fatalf("Test_InvalidRate, rate = %f, expected function to return nil settings, got %+v", val, settings)
			}
		}

	})
}

func TestWithCooldownDuration(t *testing.T) {
	t.Run("Test_WithValidCooldownDuration", func(t *testing.T) {
		name := "withCooldownDuration"
		expected := defaultSettings(name)
		duration := expected.Thresholds.CooldownDuration + 1*time.Hour
		expected.Thresholds.CooldownDuration = duration

		settings, err := circuitbreaker.NewSettings(name, circuitbreaker.WithCooldownDuration(duration))

		if err != nil {
			t.Errorf("NewSetttings(%s, circuitbreaker.CooldownDuration(%d)) expected no errors got %s", name, duration, err.Error())
		}

		compareSettings(t, expected, settings)
	})

	t.Run("Test_InvalidDuration", func(t *testing.T) {
		val := 0 * time.Hour
		settings, err := circuitbreaker.NewSettings("test", circuitbreaker.WithCooldownDuration(val))

		expectedError := circuitbreaker.ErrInvalidSettingParam{Param: "CooldownDuration", Val: val}

		if err != expectedError {
			t.Fatalf("Test_InvalidRate, duration = %d, expected error to be '%s', got '%s'", val, expectedError, err)
		}
		if settings != nil {
			t.Fatalf("Test_InvalidRate, duration = %d, expected function to return nil settings, got %+v", val, settings)
		}
	})
}

func TestWithMaxRequestOnHalfOpen(t *testing.T) {
	t.Run("Test_WithValidMaxRequestOnHalfOpen", func(t *testing.T) {
		name := "withMaxRequestOnHalfOpen"
		expected := defaultSettings(name)
		val := expected.Thresholds.MaxRequestOnHalfOpen + 100
		expected.Thresholds.MaxRequestOnHalfOpen = val

		settings, err := circuitbreaker.NewSettings(name, circuitbreaker.WithMaxRequestOnHalfOpen(val))

		if err != nil {
			t.Errorf("NewSetttings(%s, circuitbreaker.MaxRequestOnHalfOpen(%d)) expected no errors got %s", name, val, err.Error())
		}

		compareSettings(t, expected, settings)
	})

	t.Run("Test_InvalidSize", func(t *testing.T) {
		invalidValues := []int{-1, 0}

		for _, val := range invalidValues {
			settings, err := circuitbreaker.NewSettings("test", circuitbreaker.WithMaxRequestOnHalfOpen(val))

			expectedError := circuitbreaker.ErrInvalidSettingParam{Param: "MaxRequestOnHalfOpen", Val: val}

			if err != expectedError {
				t.Fatalf("Test_InvalidRate, rate = %d, expected error to be '%s', got '%s'", val, expectedError, err)
			}
			if settings != nil {
				t.Fatalf("Test_InvalidRate, rate = %d, expected function to return nil settings, got %+v", val, settings)
			}
		}

	})
}

func TestWithIsSuccessfulHandler(t *testing.T) {

	t.Run("Test_WithValidHandler", func(t *testing.T) {
		var TestHandler circuitbreaker.IsSuccessfulHandler = func(r *http.Response, err error) bool {
			return false
		}
		name := "withWindowSize"
		expected := defaultSettings(name)
		expected.IsSuccessful = TestHandler

		settings, err := circuitbreaker.NewSettings(name, circuitbreaker.WithIsSuccessfulHandler(TestHandler))

		if err != nil {
			t.Errorf("NewSetttings(%s, circuitbreaker.IsSuccessful(TestHandler)) expected no errors got %s", name, err.Error())
		}

		compareSettings(t, expected, settings)
	})

	t.Run("Test_WithNilHandler", func(t *testing.T) {
		name := "withWindowSize"
		settings, err := circuitbreaker.NewSettings(name, circuitbreaker.WithIsSuccessfulHandler(nil))

		expectedError := circuitbreaker.ErrInvalidSettingParam{Param: "IsSuccessful", Val: nil}

		if err != expectedError {
			t.Fatalf("Test_WithNilHandler, expected error to be '%s', got '%s'", expectedError, err)
		}
		if settings != nil {
			t.Fatalf("Test_WithNilHandler, expected function to return nil settings, got %+v", settings)
		}
	})
}
