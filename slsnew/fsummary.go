package main

import "fmt"
import "os"
import "bufio"
import "strings"
import "strconv"
import "time"
import "github.com/araddon/go-metrics/metrics"

// Floating point values.

type fcombiner struct {
	hist    *histogram
	entries map[string][]string
	headers []string
	start   int64
	end     int64
}

func newFcombiner() *fcombiner {
	l := new(fcombiner)
	l.entries = make(map[string][]string)
	return l
}

// Creates a summary from a set of log files.
func Fsummarize(logfiles []string) (string, error) {
	l := newFcombiner()

	err := l.proc(logfiles)
	if err != nil {
		return "", err
	}

	return l.report(), nil
}

// Reads a log file.
func (c *fcombiner) read(fileName string) error {
	fh, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	for j := 0; scanner.Scan(); j++ {
		line := scanner.Text()
		if j == 0 {
			fields := strings.Fields(string(line))
			if len(c.headers) == 0 {
				c.headers = fields[1:]
			} else if len(c.headers) > len(fields[1:]) {
				c.headers = fields[1:]
			}
		} else if j > 1 {
			fields := strings.Fields(string(line))
			c.entries[fields[0]] = fields[1:]
		}
	}

	// check for scanner errors
	return nil
}

// Process the log files and create a sorted index.
func (c *fcombiner) proc(logfiles []string) error {
	// timestamp -> data

	for _, fileName := range logfiles {
		err := c.read(fileName)
		if err != nil {
			return err
		}
	}

	rowcnt := len(c.entries)

	hist := newHistograms(c.headers, rowcnt)

	var min int64 = -1
	var max int64 = -1
	for k, entry := range c.entries {
		o, _ := strconv.ParseInt(k, 10, 64)
		if min == -1 || min > o {
			min = o
		}
		if max == -1 || max < o {
			max = o
		}
		for l, col := range c.headers {
			m, _ := strconv.ParseFloat(entry[l], 64)
			hist.update(col, m)
		}
	}

	fmt.Printf("%v %v\n", max, min)

	c.start = min
	c.end = max

	c.hist = hist

	return nil
}

func (l *fcombiner) report() string {
	var report string
	ps := [5]float64{0.50, 0.75, 0.95, 0.99, 0.999}

	report = fmt.Sprintf("col\tstart\tend\tcount\tmin\tmax\tmean\tstddev")
	for _, p := range ps {
		report += fmt.Sprintf("\t%d-precentile", int(p*100))
	}
	report += fmt.Sprintf("\n")

	start := (time.Unix(l.start, 0)).String()
	end := (time.Unix(l.end, 0)).String()

	for _, col := range l.headers {
		hist := l.hist.get(col)
		report += fmt.Sprintf("%s", col)
		report += fmt.Sprintf("\t%v", start)
		report += fmt.Sprintf("\t%v", end)
		report += fmt.Sprintf("\t%d", int(hist.GetCount()))
		report += fmt.Sprintf("\t%v", hist.GetMin())
		report += fmt.Sprintf("\t%v", hist.GetMax())
		report += fmt.Sprintf("\t%v", hist.GetMean())
		report += fmt.Sprintf("\t%v", hist.GetStdDev())
		for _, v := range hist.GetPercentiles(ps[0:]) {
			report += fmt.Sprintf("%.2f\t", v)
		}
		report += fmt.Sprintf("\n")
	}

	return report
}

type histogram struct {
	h map[string]*metrics.Histogram
	c []string
}

func (s *histogram) Add(values []float64) {
	for i := 0; i < len(s.c); i++ {
		s.h[s.c[i]].Update(values[i])
	}
}

func newHistograms(cols []string, rowcnt int) *histogram {
	hi := new(histogram)

	h := make(map[string]*metrics.Histogram, len(cols))
	for _, col := range cols {
		s := metrics.NewUniformSample(rowcnt)
		h[col] = metrics.NewHistogram(s)
	}

	hi.h = h
	hi.c = cols

	return hi
}

func (s *histogram) get(col string) *metrics.Histogram {
	return s.h[col]
}

func (s *histogram) update(col string, val float64) {
	s.h[col].Update(val)
}
