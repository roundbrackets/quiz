package main

import "io/ioutil"
import "regexp"
import "fmt"
import "os"
import "bufio"
import "strings"
import "strconv"
import "sort"
import "github.com/rcrowley/go-metrics"
import "math"

// D="../slsdir/primary/net/10a467f9158eb03510b0a99f0fab5074/";for i in `ls $D`;do grep -v "#" $D$i; done | sort | uniq | wc -l

type combiner struct {
	sample      map[string]metrics.Sample
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

func (l *combiner) initSample() {
	l.sample = make(map[string]metrics.Sample)
	size := len(l.entries)
	for _, name := range l.headers {
		l.sample[name] = metrics.NewUniformSample(size)
	}
}

// Creates a combined log file
func Combine(basedir, out string) error {
	l := newCombiner(basedir)
	l.out = out

	err := l.files()
	if err != nil {
		return err
	}

	err = l.proc()
	if err != nil {
		return err
	}

	err = l.write()
	if err != nil {
		return err
	}

	return nil
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

	l.initSample()

	hcnt := len(l.headers)

	for _, ts := range l.sortedIndex {
		for i, field := range l.entries[ts] {
			// Adjust for the smallest set.
			if i >= hcnt {
				break
			}
			name := l.headers[i]

			// It isn't clear to me how to utilize the metrics library
			// when working with floats, so I decided to truncate them.
			// I found this Float64bits(f float64) uint64, but an unsigned
			// integer is about as helpful as a float.
			value, _ := strconv.ParseFloat(field, 64)
			value = math.Trunc(value)
			l.sample[name].Update(int64(value))
		}
	}

	return l.report(), nil
}

func (c *combiner) write() error {
	f, err := os.Create(c.out)
	if nil != err {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	// c.ts contains the name of the timestamp field.
	line := c.ts + "\t" + strings.Join(c.headers, "\t") + "\n"
	_, err = w.WriteString(line)
	if nil != err {
		return err
	}
	for _, i := range c.sortedIndex {
		// Reducing to minimal column set by assuming it's the cols on the end
		// that's missing. The first col is timestamp.
		line := string(i) + "\t" + strings.Join(c.entries[i][0:len(c.headers)], "\t") + "\n"
		_, err = w.WriteString(line)
		if nil != err {
			return err
		}
	}
	w.Flush()
	return nil
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

	for col, sample := range l.sample {
		report += fmt.Sprintf("%s", col)
		report += fmt.Sprintf("\t%d - %d", start, end)

		report += fmt.Sprintf("\t%d", sample.Size())
		report += fmt.Sprintf("\t%d", sample.Min())
		report += fmt.Sprintf("\t%d", sample.Max())
		// The values for mean and std dev are unweildy, but I don't know
		// enough round them properly.
		report += fmt.Sprintf("\t%v", sample.Mean())
		report += fmt.Sprintf("\t%v", sample.StdDev())
		for _, v := range sample.Percentiles(ps[0:]) {
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
