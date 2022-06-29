package worker

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

type Result struct {
	Line       string
	LineNumber int
	Path       string
	Index      [][]int
}

type Results struct {
	Inner []Result
}

var (
	Info = Teal
	Warn = Yellow
	Fata = Red
)

var (
	Black   = Color("\033[1;30m%s\033[0m")
	Red     = Color("\033[1;31m%s\033[0m")
	Green   = Color("\033[1;32m%s\033[0m")
	Yellow  = Color("\033[1;33m%s\033[0m")
	Purple  = Color("\033[1;34m%s\033[0m")
	Magenta = Color("\033[1;35m%s\033[0m")
	Teal    = Color("\033[1;36m%s\033[0m")
	White   = Color("\033[1;37m%s\033[0m")
)

func (r *Result) Print() string {
	sl, s, sr := substr(r.Line, r.Index[0][0], r.Index[0][1]-r.Index[0][0]+1)
	return fmt.Sprintf("%v:%v: %v%v%v", r.Path, r.LineNumber, sl, Red(s), sr)
}

func substr(input string, start int, length int) (sl, s, sr string) {
	asRunes := []rune(input)

	if start >= len(asRunes) {
		return "", "", ""
	}

	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}

	sl = string(asRunes[0:start])
	s = string(asRunes[start : start+length])
	sr = string(asRunes[start+length:])
	return
}

func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}

func NewResult(line string, lineNumber int, path string, index [][]int) Result {
	return Result{line, lineNumber, path, index}
}

func FindInFile(path string, find string) *Results {
	regexpPattern, _ := regexp.Compile(find)
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening file: %v", err)
		return nil
	}

	results := Results{make([]Result, 0)}
	scanner := bufio.NewScanner(file)
	lineNumber := 1
	for scanner.Scan() {
		line := scanner.Text()
		if regexpPattern.MatchString(line) {
			index := regexpPattern.FindAllStringSubmatchIndex(line, -1)
			results.Inner = append(results.Inner, NewResult(line, lineNumber, path, index))
		}
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v", err)
		return nil
	}

	if len(results.Inner) > 0 {
		return &results
	} else {
		return nil
	}
}
