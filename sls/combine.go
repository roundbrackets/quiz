package main

import "io/ioutil"
import "regexp"
import "fmt"
import "os"
import "bufio"
import "strings"
import "strconv"
import "sort"

// D="../slsdir/primary/net/10a467f9158eb03510b0a99f0fab5074/";for i in `ls $D`;do grep -v "#" $D$i; done | sort | uniq | wc -l
// 147

type combiner struct {
	entries     map[int]string
	sortedIndex []int
	headers     []string
	logfiles    []string
    basedir     string
    out         string
}

type Combined struct {
    Filename string
    Headers []string
    RecCnt int
    First int
    Last int
}

func newCombiner (basedir string, out string) *combiner {
    l := new (combiner)
    l.basedir = basedir
    l.out = out
    return l
}

func Combine (basedir, out string) (*Combined, error) {
    l := newCombiner(basedir, out)

	err := l.files()
	if err != nil {
        return nil, err
	}

	err = l.proc()
	if err != nil {
        return nil, err
	}

	err = l.write()
	if err != nil {
        return nil, err
	}

    combined := new (Combined)
    combined.Filename = l.out
    combined.Headers = l.headers
    combined.RecCnt = len(l.entries)
    combined.First = l.sortedIndex[0]
    combined.Last = l.sortedIndex[len(l.sortedIndex)-1]

    return combined, nil
}

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

func (c *combiner) proc() error {
    c.entries = make(map[int]string)

	for _, fileName := range c.logfiles {
        err := c.read(fileName)
        if err != nil {
            return err
        }
	}

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

func (c *combiner) write() error {
	f, err := os.Create(c.out)
	if nil != err {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, i := range c.sortedIndex {
        //_, err = w.WriteString(fmt.Sprintf("%d\t%s\n", i, strings.Join(data.entries[i], "\t")))
        line := fmt.Sprintf("%s\n", c.entries[i])
        //ne[len(fields[0])+1:]fmt.Printf("%s %v\n", line, i)
        _, err = w.WriteString(line)
		if nil != err {
            // delete file?
			return err
		}
	}
	w.Flush()
    return nil
}

func (c *combiner) read (fileName string) error {
        //fmt.Printf("Filename %s\n", fileName)

		fh, err := os.Open(fileName)
        if err != nil {
            return err
        }
        defer fh.Close()

		scanner := bufio.NewScanner(fh)
		for j := 0; scanner.Scan(); j++ {
			line := scanner.Text()
            //fmt.Print(line)
			if j == 0 {
				fields := strings.Fields(string(line))
				c.headers = fields[2:]
			} else if j > 1 {
				fields := strings.Fields(line)
				ts, _ := strconv.ParseInt(fields[0], 10, 64)
                c.entries[int(ts)] = line[len(fields[0])+1:]
                //fmt.Printf("%s %v\n", int(ts), line[len(fields[0])+1:])
                //strings.Join(fields[1:], "\t")
			}
		}
        // check for scanner errors
        return nil
}
