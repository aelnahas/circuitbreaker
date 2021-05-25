package gauges

type FixedWindowGauge struct {
	windowSize     int
	head           int
	measurements   []*Aggregate
	totalAggregate *Aggregate
}

var _ Gauge = &FixedWindowGauge{}

func NewFixedWindowGauge(windowSize int) *FixedWindowGauge {
	gauge := &FixedWindowGauge{
		windowSize:     windowSize,
		head:           0,
		totalAggregate: &Aggregate{},
	}

	gauge.makeNewMeasurements()
	return gauge
}

func (g *FixedWindowGauge) LogReading(outcome Outcome) {
	g.slideWindow()
	measurement := g.measurements[g.head]
	measurement.record(outcome)
	g.totalAggregate.record(outcome)
}

func (g *FixedWindowGauge) OverallAggregate() Aggregate {
	return *g.totalAggregate
}

func (g *FixedWindowGauge) LatestMeasurement() (Aggregate, error) {

	if g.totalAggregate.RequestCount == 0 {
		return Aggregate{}, ErrEmptyMeasurements
	}
	return *g.measurements[g.head], nil
}

func (g *FixedWindowGauge) Reset() {
	g.head = 0
	g.totalAggregate = &Aggregate{}
	g.makeNewMeasurements()
}

func (g *FixedWindowGauge) slideWindow() {
	tail := g.tail()
	oldMeasurement := g.measurements[tail]
	g.totalAggregate.erase(oldMeasurement)
	oldMeasurement.reset()
	g.head = tail
}

func (g *FixedWindowGauge) tail() int {
	return (g.head + 1) % g.windowSize
}

func (g *FixedWindowGauge) makeNewMeasurements() {
	g.measurements = make([]*Aggregate, g.windowSize)

	for i := 0; i < g.windowSize; i++ {
		g.measurements[i] = &Aggregate{}
	}
}
