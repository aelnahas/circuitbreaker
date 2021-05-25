package gauges_test

import (
	"fmt"
	"testing"

	"github.com/aelnahas/circuitbreaker/circuitbreaker/gauges"
)

func TestFixedWindowGauge(t *testing.T) {
	t.Run("TestReset", func(t *testing.T) {
		gauge := gauges.NewFixedWindowGauge(3)
		gauge.LogReading(gauges.Success)
		gauge.LogReading(gauges.Failure)
		gauge.Reset()

		expectedAggregate := gauges.Aggregate{}
		aggregate := gauge.OverallAggregate()

		if aggregate != expectedAggregate {
			t.Errorf("Aggregate expected %+v got %+v", expectedAggregate, aggregate)
		}
	})

	t.Run("TestLatestMeasurment", func(t *testing.T) {
		t.Run("WhenEmpty", func(t *testing.T) {
			gauge := gauges.NewFixedWindowGauge(3)
			_, err := gauge.LatestMeasurement()
			if err != gauges.ErrEmptyMeasurements {
				t.Errorf("TestFixedWindowGauge.TestLatestMeasurement/WhenEmpty Error expected ErrEmptyMeasurements got %s", err)
			}
		})

		t.Run("WhenMeasurmentsCountIsLessThanWindow", func(t *testing.T) {
			windowSize := 3
			gauge := gauges.NewFixedWindowGauge(windowSize)

			gauge.LogReading(gauges.Success)
			gauge.LogReading(gauges.Failure)

			measurement, err := gauge.LatestMeasurement()
			if err != nil {
				t.Errorf("Error expected nil got %s", err)
			}

			fmt.Printf("%+v\n", measurement)
			expectedMeasurement := gauges.Aggregate{
				RequestCount: 1,
				SuccessCount: 0,
				FailureCount: 1,
			}

			if measurement != expectedMeasurement {
				t.Errorf("Measurement expected %+v got %+v", expectedMeasurement, measurement)
			}
		})

		t.Run("WhenMeasurmentsCountIsMoreThanWindow", func(t *testing.T) {
			windowSize := 3
			gauge := gauges.NewFixedWindowGauge(windowSize)

			gauge.LogReading(gauges.Failure)
			gauge.LogReading(gauges.Failure)
			gauge.LogReading(gauges.Failure)
			gauge.LogReading(gauges.Success)

			measurement, err := gauge.LatestMeasurement()
			if err != nil {
				t.Errorf("Error expected nil got %s", err)
			}

			expectedMeasurement := gauges.Aggregate{
				RequestCount: 1,
				SuccessCount: 1,
				FailureCount: 0,
			}

			if measurement != expectedMeasurement {
				t.Errorf("Measurement expected %+v got %+v", expectedMeasurement, measurement)
			}
		})
	})

	t.Run("TestOverallAggregate", func(t *testing.T) {
		windowSize := 3
		gauge := gauges.NewFixedWindowGauge(windowSize)

		for i := 0; i < windowSize; i++ {
			gauge.LogReading(gauges.Success)
		}

		for i := 0; i < windowSize; i++ {
			gauge.LogReading(gauges.Failure)
		}

		aggregate := gauge.OverallAggregate()
		expectedAggregate := gauges.Aggregate{RequestCount: windowSize, FailureCount: windowSize, SuccessCount: 0}

		if aggregate != expectedAggregate {
			t.Errorf("OverallAggregate expected %+v got %+v", expectedAggregate, aggregate)
		}
	})
}
