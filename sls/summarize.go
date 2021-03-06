package main

import "fmt"
import "github.com/rcrowley/go-metrics"
import "os"
import "bufio"
import "strings"
import "strconv"
import "math"
import "io/ioutil"
import "regexp"

const BASEDIR = "../slsdir/primary/net/"

type Summary struct {
    sample      map[string]metrics.Sample
    data        map[string][]int64
    log         *Combined
}

func NewSummary (log *Combined) *Summary {
    summary := new (Summary)
    summary.log = log
    return summary
}

func (l *Summary) initSample () {
    l.sample = make(map[string]metrics.Sample);
    l.data = make(map[string][]int64);
    for _, name := range l.log.Headers {
        l.sample[name] = metrics.NewUniformSample(l.log.RecCnt)
        l.data[name] = make([]int64, l.log.RecCnt);
    }
}

func (l *Summary) Summarize () {
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
            //func Float64bits(f float64) uint64
            value, _ := strconv.ParseFloat(field, 64)
            value = math.Trunc(value)
            //fmt.Printf("%s %v\n", name, value);
            l.sample[name].Update(int64(value))
            l.data[name][j] = int64(value)
        }
    }
}

func (l *Summary) Report () {
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
        fmt.Printf("\t%v", sample.Mean())
        fmt.Printf("\t%v", sample.StdDev())
        for _, v := range sample.Percentiles(ps[0:]) {
           fmt.Printf("%.2f\t", v)
        }
        fmt.Printf("\n")
    }
}

func main() {
    dir, _ := InDir(BASEDIR)
    out := fmt.Sprintf("%s/combined_logs", dir)

    log, err := Combine(dir, out)

    if err != nil {
        fmt.Printf("%+v %T", err)
        return
    }

    summary := NewSummary(log)

    summary.Summarize ()
    summary.Report ()
}

func InDir (basedir string) (string, error) {
	id, err := NodeId(basedir)
	if err == nil {
	    dir := fmt.Sprintf("%s%s", basedir, id)
        return dir, nil
	}
	return "", err
}

func NodeId(basedir string) (string, error) {
	files, err := ioutil.ReadDir(basedir)

	if err == nil {
		for _, fi := range files {
			if fi.IsDir() {
				match, err := regexp.MatchString("^[0-9A-Fa-f]+$", fi.Name())
				if err != nil {
					break
				}
				if match {
					return fi.Name(), nil
				}
			}
		}
	}

	return "", err
}
