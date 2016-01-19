/*
* Q: given a list of words like http://norvig.com/ngrams/word.list find the 
* longest word, which is itself composed of words which must be in the wordlist 
* input. The coding shouldn't take more than one hour. Any programming language 
* can be used.
*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const WORDLIST = "./word.list"

type Words struct {
	master   []string
	byLength map[int][]int
	alpha    map[string][]int
	min      int
	max      int
	cnt      int
}

func main() {
	words, err := newWords()
	if err != nil {
		fmt.Println("%s", err)
		return
	}

	var foundWord string
	wordLength := 0

Loop:
	for i := words.max; i >= words.min; i-- {
		list, err := words.getByLength(i)
		if nil != err {
			continue
		}

		for _, word := range list {
			startWord := words.w(word)
			//fmt.Printf("Starting %s(%d)...\n", startWord, i)
			subwords := words.subWords(startWord)

			//fmt.Printf("Subwords: %v\n", subwords)

			for _, sword := range subwords {
				if word == sword {
					continue
				}

				subWord := words.w(sword)
				nextWord := startWord[len(subWord):]

				//fmt.Printf("subWord %s Word %s NextWord %s\n", subWord,
				//startWord, nextWord)

				if follow_paths(nextWord, words) {
					//fmt.Printf("Match %s\n", nextWord)
					foundWord = words.w(word)
					wordLength = len(startWord)
				}

				if wordLength > 0 {
					break Loop
				}
			}
		}
	}

	if wordLength > 0 {
		fmt.Printf("Found '%s' of length %d.\n", foundWord, wordLength)
	} else {
		fmt.Printf("Found no word.\n")
	}
}

func follow_paths(word string, words *Words) bool {
	if len(word) < words.min {
		return false
	}

	subwords := words.subWords(word)

	//fmt.Printf("Subwords: %v\n", subwords)

	for _, w := range subwords {
		if words.w(w) == word {
			//fmt.Printf("Match %s\n", word)
			return true
		}
	}

	for _, w := range subwords {
		subWord := words.w(w)
		nextWord := word[len(subWord):]

		//fmt.Printf("subWord %s Word %s NextWord %s\n", subWord, word,
		//nextWord)

		if follow_paths(nextWord, words) {
			//fmt.Printf("Match %s\n", nextWord)
			return true
		}
	}

	return false
}

func newWords() (*Words, error) {
	fh, err := os.Open(WORDLIST)
	defer fh.Close()
	if err != nil {
		return nil, err
	}

	words := new(Words)

	length := listlen()
	master := make([]string, length)

	byLength := make(map[int][]int)
	alpha := make(map[string][]int)

	min := length
	max := 0
	scanner := bufio.NewScanner(fh)
	for i := 0; scanner.Scan(); i++ {
		master[i] = scanner.Text()
		l := len(master[i]) // length

		if _, OK := byLength[l]; !OK {
			byLength[l] = make([]int, 0, length)
		}

		c := fmt.Sprintf("%s%d", string(master[i][0]), l)
		if _, OK := alpha[c]; !OK {
			alpha[c] = make([]int, 0, length)
		}

		byLength[l] = append(byLength[l], i)
		alpha[c] = append(alpha[c], i)

		if l < min {
			min = l
		}
		if l > max {
			max = l
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	words.master = master
	words.byLength = byLength
	words.alpha = alpha
	words.min = min
	words.max = max
	words.cnt = length

	return words, err
}

func (w *Words) getByLength(length int) ([]int, error) {
	if arr, OK := w.byLength[length]; OK {
		return arr, nil
	}
	return nil, fmt.Errorf("No such index.")
}

func (w *Words) getByAlphaLength(first byte, length int) ([]int, error) {
	c := fmt.Sprintf("%s%d", string(first), length)
	if _, OK := w.alpha[c]; OK {
		return w.alpha[c], nil
	}
	return nil, fmt.Errorf("No such index, %s.\n", c)
}

func (w *Words) w(index int) string {
	return w.master[index]
}

func (w *Words) subWords(word string) []int {
	found := make([]int, 0, w.cnt)
	first := word[0]
	for i := w.min; i <= len(word); i++ {
		words, err := w.getByAlphaLength(first, i)
		//fmt.Printf("%s words: %v\n", string(first), words);
		if err != nil {
			continue
		}
		//fmt.Printf("Starting subwords with %s %d cnt %d.\n", string(first),
		//i, len(words));
		for _, ipos := range words {
			pos := w.w(ipos)
			//fmt.Printf("Comparing %s and %s.\n", word, pos);
			f := 1
			for j := 0; j < len(pos); j++ {
				if pos[j] != word[j] {
					f = 0
					break
				}
			}
			if f == 1 {
				found = append(found, ipos)
			}
		}
	}
	return found
}

func (w *Words) checkAlpha() {
	for i, arr := range w.alpha {
		fmt.Printf("%s: ", i)
		for _, word := range arr {
			fmt.Printf("%s ", w.w(word))
		}
		fmt.Printf("\n")
	}
}

func (w *Words) checkByLength() {
	for i, arr := range w.byLength {
		for j, word := range arr {
			fmt.Printf("|%d %d %s|\n", i, j, w.w(word))
		}
	}
}

func listlen() int {
	size := 1000

	wc, err := exec.Command("wc", "-l", WORDLIST).Output()
	if nil == err {
		wc := string(wc)
		fwc := strings.Fields(wc)

		if len(fwc) == 2 {
			lwc, err := strconv.ParseInt(fwc[0], 10, 64)
			if nil == err {
				size = int(lwc)
			}
		}
	}

	return size
}
