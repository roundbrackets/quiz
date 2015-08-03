package main

import "fmt"
import "io/ioutil"
import "regexp"

const BASEDIR = "../slsdir/primary/net/"

func main() {
	dir, err := InDir(BASEDIR)
	if err != nil {
		fmt.Println(err)
		return
	}
	files, err := Files(dir)
	if err != nil {
		fmt.Println(err)
		return
	}
	out := fmt.Sprintf("%s/combined_logs", dir)

	// Write to a file
	err = Combine(files, out)
	if err != nil {
		fmt.Println(err)
		return
	}

	log, err := ioutil.ReadFile(out)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(out)
	fmt.Println(string(log))

	// Or, make a summary
	data, err := Summarize(files)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(data)

	// Or, make a float summary
	data, err = Fsummarize(files)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(data)
}

func InDir(basedir string) (string, error) {
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

// Find log files to process.
func Files(basedir string) ([]string, error) {
	files, err := ioutil.ReadDir(basedir)
	logfiles := make([]string, 0, len(files))

	if err == nil {
		for _, fi := range files {
			if !fi.IsDir() {
				match, err := regexp.MatchString("^[0-9]+.*$", fi.Name())
				if err != nil {
					return nil, err
				}
				if match {
					logfiles = append(logfiles, fmt.Sprintf("%s/%s", basedir, fi.Name()))
				}
			}
		}
	}

	if len(logfiles) == 0 {
		return nil, fmt.Errorf("No logfiles found.")
	}

	return logfiles, err
}
