package fsummary

import "io/ioutil"
import "regexp"
import "fmt"
import "os"
import "bufio"
import "strings"
import "strconv"
import "sort"
import "github.com/araddon/go-metrics/metrics"

// D="../slsdir/primary/net/10a467f9158eb03510b0a99f0fab5074/";for i in `ls $D`;do grep -v "#" $D$i; done | sort | uniq | wc -l

type combiner struct {
	hist      map[string]*metrics.Histogram
	entries     map[int][]string
	sortedIndex []int
	headers     []string
	logfiles    []string
	basedir     string
	out         string
	ts          string
}

func newCombiner(basedir string) *combiner {
	l := new(combiner)
	l.basedir = basedir
	return l
}

func (l *combiner) init() {
	l.hist = make(map[string]*metrics.Histogram)
	size := len(l.entries)
	for _, name := range l.headers {
        h := metrics.NewUniformSample(size)
		l.hist[name] = metrics.NewHistogram(h)
	}
}

// Creates a summary from a set of log files.
func Summarize(basedir string) (string, error) {
	l := newCombiner(basedir)

	err := l.files()
	if err != nil {
		return "", err
	}

	err = l.proc()
	if err != nil {
		return "", err
	}

	l.init()

	hcnt := len(l.headers)

	for _, ts := range l.sortedIndex {
		for i, field := range l.entries[ts] {
			// Adjust for the smallest set.
			if i >= hcnt {
				break
			}
			name := l.headers[i]

			value, _ := strconv.ParseFloat(field, 64)
			l.hist[name].Update(value)
		}
	}

	return l.report(), nil
}

// Reads a log file.
func (c *combiner) read(fileName string) error {
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
				c.ts = fields[0]
				c.headers = fields[1:]
			} else if len(c.headers) > len(fields[1:]) {
				c.headers = fields[1:]
			}
		} else if j > 1 {
			fields := strings.Fields(line)
			ts, _ := strconv.ParseInt(fields[0], 10, 64)
			c.entries[int(ts)] = fields[1:]
		}
	}
	// check for scanner errors
	return nil
}

func (l *combiner) report() string {
	var report string
	ps := [5]float64{0.50, 0.75, 0.95, 0.99, 0.999}

	report = fmt.Sprintf("col\tperiod\tcount\tmin\tmax\tmean\tstddev")
	for _, p := range ps {
		report += fmt.Sprintf("\t%d-precentile", int(p*100))
	}
	report += fmt.Sprintf("\n")
	start := l.sortedIndex[0]
	end := l.sortedIndex[len(l.sortedIndex)-1]

	for col, s := range l.hist {
		report += fmt.Sprintf("%s", col)
		report += fmt.Sprintf("\t%d - %d", start, end)
		report += fmt.Sprintf("\t%v", len(l.entries[l.sortedIndex[0]]))
		report += fmt.Sprintf("\t%v", s.GetMin())
		report += fmt.Sprintf("\t%v", s.GetMax())
		// The values for mean and std dev are unweildy, but I don't know
		// enough round them properly.
		report += fmt.Sprintf("\t%v", s.GetMean())
		report += fmt.Sprintf("\t%v", s.GetStdDev())
		for _, v := range s.GetPercentiles(ps[0:]) {
			report += fmt.Sprintf("%.2f\t", v)
		}
		report += fmt.Sprintf("\n")
	}

	return report
}

// Find log files to process.
func (c *combiner) files() error {
	files, err := ioutil.ReadDir(c.basedir)
	c.logfiles = make([]string, 0, len(files))

	if err == nil {
		for _, fi := range files {
			if !fi.IsDir() {
				match, err := regexp.MatchString("^[0-9]+.*$", fi.Name())
				if err != nil {
					return err
				}
				if match {
					c.logfiles = append(c.logfiles, fmt.Sprintf("%s/%s", c.basedir, fi.Name()))
				}
			}
		}
	}

	return err
}

// Process the log files and create a sorted index.
func (c *combiner) proc() error {
	// timestamp -> data
	c.entries = make(map[int][]string)

	for _, fileName := range c.logfiles {
		err := c.read(fileName)
		if err != nil {
			return err
		}
	}

	// Make a index sorted by timestamp.
	keys := make([]int, len(c.entries))
	i := 0
	for k, _ := range c.entries {
		keys[i] = k
		i++
	}
	sort.Ints(keys)
	c.sortedIndex = keys

	return nil
}

// Write data to a file.
