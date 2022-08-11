package p2c

import (
	"github.com/wcharczuk/go-chart"
	"math"
	"math/rand"
	"os"
)

func main() {
	xAxis, yAxis := make([]float64, 1000), make([]float64, 1000)
	for i := 0;i < 1000;i++ {
		xAxis[i] = float64(i+1)
	}
	for i := 0;i < 1000;i++ {
		var latency float64
		if i >= 100 && i <= 200 {
			latency = 125
		} else {
			latency = float64(rand.Intn(25))
		}
		yAxis[i] = calEWMA(latency)
	}
	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name: "req",
		},
		YAxis: chart.YAxis{
			Name: "latency",
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				//Style: chart.Style{
				//	StrokeColor: chart.GetDefaultColor(0).WithAlpha(64),
				//	FillColor:   chart.GetDefaultColor(0).WithAlpha(64),
				//},
				XValues: xAxis,
				YValues: yAxis,
			},
		},
	}
	f, _ := os.Create("output.png")
	defer f.Close()
	_ = graph.Render(chart.PNG, f)
}

var totalLatency,times float64
func calAvg(latency float64) float64 {
	totalLatency += latency
	times++
	return totalLatency / times
}

const Weight = 0.95
var lastEWMA float64
func calEWMA(latency float64) float64 {
	weight := calWeight(latency)
	lastEWMA = weight*lastEWMA + (1 - weight)*latency
	return lastEWMA
}

const K = 800
func calWeight(latency float64) float64 {
	return math.Exp(-latency/K)
}
