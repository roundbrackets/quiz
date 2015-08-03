package main

import "fmt"
import "io/ioutil"
import "regexp"

const BASEDIR = "../slsdir/primary/net/"

func main() {
	dir, _ := InDir(BASEDIR)
	out := fmt.Sprintf("%s/combined_logs", dir)

	log, err := Combine(dir, out)
	if err != nil {
		fmt.Printf("%+v %T", err)
		return
	}

	summary := NewSummary(log)
	summary.Summarize()
	summary.Report()
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
