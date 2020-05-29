package dudect

import (
	"fmt"
	"github.com/dterei/gotsc"
	"github.com/montanaflynn/stats"
	"math"
)

const (
	enoughMeasurements = 10000
	numberPercentiles  = 100
	numberTests        = 1 + numberPercentiles + 1
	tThresholdBananas  = 500
	tThresholdModerate = 10
)

type testData struct {
	mean [2]float64
	m2   [2]float64
	n    [2]float64
}

// Input is the data structure representing the input data and its categorization.
type Input struct {
	// Data is the actual input fed into the computation function.
	Data []byte
	// Class is the categorization of the input, and must be 0 or 1.
	Class uint8
}

func initTestData() testData {
	t := testData{}
	t.mean = [2]float64{0.0, 0.0}
	t.m2 = [2]float64{0.0, 0.0}
	t.n = [2]float64{0, 0}
	return t
}

func (t *testData) push(newData float64, class uint8) {
	if class != 0 && class != 1 {
		panic(fmt.Sprintf("attempt to push Data of Class %d != (1 or 0)", class))
	}
	t.n[class] += 1
	delta := newData - t.mean[class]
	t.mean[class] = t.mean[class] + delta/t.n[class]
	t.m2[class] = t.m2[class] + delta*(newData-t.mean[class])
}

func (t *testData) compute() float64 {
	variance := [2]float64{0, 0}
	variance[0] = t.m2[0] / (t.n[0] - 1)
	variance[1] = t.m2[1] / (t.n[1] - 1)
	num := t.mean[0] - t.mean[1]
	den := math.Pow(variance[0]/t.n[0]+variance[1]/t.n[1], 0.5)
	return num / den // t_value
}

func maxTest(t []testData) int {
	testID := 0
	max := float64(0)
	for i := range t {
		if t[i].n[0]+t[i].n[1] > enoughMeasurements {
			currentT := math.Abs(t[i].compute())
			if currentT > max {
				max = currentT
				testID = i
			}
		}
	}
	return testID
}

func report(t []testData) {
	mt := maxTest(t)
	maxT := math.Abs(t[mt].compute())
	maxTN := t[mt].n[0] + t[mt].n[1]
	maxTau := maxT / math.Sqrt(maxTN)
	fmt.Printf("total measurements: %7.2f Million\n", maxTN/1e6)
	fmt.Printf("max t-value: %7.2f, max tau: %.2e, (5/tau)^2: %.2e\n", maxT, maxTau, math.Pow(5/maxTau, 2))
	if maxT > tThresholdBananas {
		fmt.Println("Definitely not constant time.")
		return
	}
	if maxT > tThresholdModerate {
		fmt.Println("Probably not constant time.")
		return
	}
	fmt.Println("For the moment, maybe constant time.")
}

func updateStatics(measurements []float64, inputs []Input) []testData {
	percentiles := preparePercentiles(measurements)
	var t = make([]testData, numberTests)
	for i := 0; i < numberTests; i++ {
		t[i] = initTestData()
	}
	for i := range measurements {
		data := measurements[i]
		class := inputs[i].Class
		if data <= 0 {
			panic(fmt.Sprintf("Interger overflow may happens (%v)!", data))
		}
		t[0].push(data, class)

		for j := range percentiles {
			if data < percentiles[j] {
				t[j+1].push(data, class)
			}
		}

		if t[0].n[0] > enoughMeasurements {
			centered := data - t[0].mean[class]
			t[numberTests-1].push(math.Pow(centered, 2), class)
		}
	}
	return t
}

func preparePercentiles(measurements []float64) []float64 {
	var percentiles []float64
	for i := 0; i < numberPercentiles; i++ {
		p, err := stats.Percentile(measurements, (1-math.Pow(0.5, 10*float64(i+1)/numberPercentiles))*100)
		if err != nil {
			panic(fmt.Sprintf("%v exponent: %v", err, 10*float64(i+1)/numberPercentiles))
		}
		percentiles = append(percentiles, p)
	}
	return percentiles
}

func doMeasurement(init func() func([]byte), inputs []Input) []float64 {
	numberMeasurements := len(inputs)
	var measurements []float64
	doOneComputation := init()
	tscOverhead := gotsc.TSCOverhead()
	for i := 0; i < numberMeasurements; i++ {
		start := gotsc.BenchStart()
		doOneComputation(inputs[i].Data)
		end := gotsc.BenchEnd()
		measurements = append(measurements, float64(end-start-tscOverhead))
	}
	return measurements
}

// Dudect tests if the computation function returned by initState is constant time
// against two classes of inputs returned by prepareInputs.
// initState: a function returns a closure function as the target computation to be
// measured (note this function should take []byte as input
// prepareInputs: a function returns a list of Input to be fed into the computation
// function
func Dudect(initState func() func([]byte), prepareInputs func() []Input) {
	inputs := prepareInputs()
	measurements := doMeasurement(initState, inputs)
	t := updateStatics(measurements, inputs)
	report(t)
}
