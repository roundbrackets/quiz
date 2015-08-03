package main

import "fmt"
import "github.com/rcrowley/go-metrics"
import "os"
import "bufio"
import "strings"
import "strconv"
import "math"

type Summary struct {
	sample map[string]metrics.Sample
	// I started using metrics.Sample, then realized I could have used a map of
	// arrays, I'd just have to type more.
	//data   map[string][]int64
	log *Combined
}

func NewSummary(log *Combined) *Summary {
	summary := new(Summary)
	summary.log = log
	return summary
}

func (l *Summary) initSample() {
	l.sample = make(map[string]metrics.Sample)
	//l.data = make(map[string][]int64)
	for _, name := range l.log.Headers {
		l.sample[name] = metrics.NewUniformSample(l.log.RecCnt)
		//l.data[name] = make([]int64, l.log.RecCnt)
	}
}

// You could, rather than write the intermidiate file, have the combiner return
// the combined data.
func (l *Summary) Summarize() {
	l.initSample()

	fh, _ := os.Open(l.log.Filename)
	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	for j := 0; scanner.Scan(); j++ {
		line := scanner.Text()
		fields := strings.Fields(line)
		for i, field := range fields {
			if i >= len(l.log.Headers) {
				continue
			}
			name := l.log.Headers[i]

			// It isn't clear to me how to utilize the metrics library
			// when working with floats, so I decided to truncate them.
			// I found this Float64bits(f float64) uint64, but an unsigned
			// integer is about as helpful as a float.
			value, _ := strconv.ParseFloat(field, 64)
			value = math.Trunc(value)
			l.sample[name].Update(int64(value))
			//l.data[name][j] = int64(value)
		}
	}
}

func (l *Summary) Report() {
	ps := [5]float64{0.50, 0.75, 0.95, 0.99, 0.999}

	fmt.Printf("col\tperiod\tcount\tmin\tmax\tmean\tstddev")
	for _, p := range ps {
		fmt.Printf("\t%d-precentile", int(p*100))
	}
	fmt.Printf("\n")

	for col, sample := range l.sample {
		fmt.Printf("%s", col)
		fmt.Printf("\t%d - %d", l.log.First, l.log.Last)

		fmt.Printf("\t%d", sample.Size())
		fmt.Printf("\t%d", sample.Min())
		fmt.Printf("\t%d", sample.Max())
		// The values for mean and std dev seem unweildy, but I don't know
		// enough round them properly.
		fmt.Printf("\t%v", sample.Mean())
		fmt.Printf("\t%v", sample.StdDev())
		//fmt.Printf("\t%v", metrics.SampleStdDev(data[col]))
		for _, v := range sample.Percentiles(ps[0:]) {
			fmt.Printf("%.2f\t", v)
		}
		fmt.Printf("\n")
	}
}
