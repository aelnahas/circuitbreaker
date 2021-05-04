package circuitbreaker

type Outcome int

const (
	Succeeded Outcome = iota
	Failed
)

func (o Outcome) String() string {
	switch o {
	case Succeeded:
		return "succeeded"
	case Failed:
		return "failed"
	default:
		return "unknown"
	}
}

type MetricsRecorder interface {
	LastRecord() (Outcome, error)
	Record(Outcome) Metrics
	Metrics() Metrics
	Reset()
}

type Measurement struct {
	Outcome Outcome
}

type Metrics struct {
	RequestCount int
	FailureRate  float64
	FailureCount int
	SuccessRate  float64
	SuccessCount int
}

type FixedSizeSlidingWindowMetrics struct {
	WindowSize   int
	measurements []Measurement
	head         int
	metrics      Metrics
}

var _ MetricsRecorder = &FixedSizeSlidingWindowMetrics{}

func NewFixedSizeSlidingWindowMetrics(size int) *FixedSizeSlidingWindowMetrics {
	return &FixedSizeSlidingWindowMetrics{
		WindowSize:   size,
		measurements: make([]Measurement, size),
		head:         0,
		metrics:      Metrics{},
	}
}

func (fs *FixedSizeSlidingWindowMetrics) LastRecord() (Outcome, error) {
	if fs.metrics.RequestCount == 0 {
		return Failed, ErrEmptyMeasurements
	}

	return fs.measurements[fs.head].Outcome, nil
}

func (fs *FixedSizeSlidingWindowMetrics) Record(outcome Outcome) Metrics {
	fs.slideWindow()
	fs.measurements[fs.head].Outcome = outcome

	if fs.metrics.RequestCount < fs.WindowSize {
		fs.metrics.RequestCount++
	}

	if outcome == Succeeded {
		fs.metrics.SuccessCount++
	} else {
		fs.metrics.FailureCount++
	}

	fs.updateMetrics()
	return fs.metrics
}

func (fs *FixedSizeSlidingWindowMetrics) Reset() {
	fs.measurements = make([]Measurement, fs.WindowSize)
	fs.head = 0
	fs.metrics.FailureCount = 0
	fs.metrics.FailureRate = 0
	fs.metrics.RequestCount = 0
	fs.metrics.SuccessCount = 0
	fs.metrics.SuccessRate = 0
}

func (fs *FixedSizeSlidingWindowMetrics) Metrics() Metrics {
	return fs.metrics
}

func (fs *FixedSizeSlidingWindowMetrics) slideWindow() {
	fs.head = fs.tail()
	outcome := fs.measurements[fs.head].Outcome
	if fs.metrics.RequestCount >= fs.WindowSize {
		if outcome == Succeeded {
			fs.metrics.SuccessCount--
		} else {
			fs.metrics.FailureCount--
		}
	}
}

func (fs *FixedSizeSlidingWindowMetrics) tail() int {
	return (fs.head + 1) % fs.WindowSize
}

func (fs *FixedSizeSlidingWindowMetrics) updateMetrics() {
	if fs.metrics.RequestCount > 0 {
		fs.metrics.FailureRate = 100 * float64(fs.metrics.FailureCount) / float64(fs.WindowSize)
		fs.metrics.SuccessRate = 100 * float64(fs.metrics.SuccessCount) / float64(fs.WindowSize)
	}

}
