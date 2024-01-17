package searcher

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

type StringSearcher interface {
	MakeSearch(input string) error
	PrintSearchResults()
}

type searchResult struct {
	sourceName       string
	numOfOccurrences int
}

type StrSearcher struct {
	target string
	// для хранения результатов выбран слайс, потому что в тексте задания в примере работы программы вводятся
	// одинаковые адреса и для каждого отдельно пишется результат и суммируется он тоже в том числе для одинаковых
	// адресов.
	searchResults  []searchResult
	mu             *sync.Mutex
	workerChan     chan struct{}
	sumOccurrences int
}

func NewStrSearcher(target string, maxGoroutinesNum int) *StrSearcher {
	return &StrSearcher{
		target:        target,
		searchResults: make([]searchResult, 0),
		mu:            &sync.Mutex{},
		workerChan:    make(chan struct{}, maxGoroutinesNum),
	}
}

func (s *StrSearcher) MakeSearch(input string) error {
	s.workerChan <- struct{}{}
	defer func() {
		<-s.workerChan
	}()
	if s.isValidURL(input) {
		res, err := s.searchHTTP(input)
		if err == nil {
			s.addResult(input, res)
		}
		return err
	}
	if s.isValidFile(input) {
		res, err := s.searchInFile(input)
		if err == nil {
			s.addResult(input, res)
		}
		return err
	}
	return fmt.Errorf("invalid input")
}

func (s *StrSearcher) PrintSearchResults() {
	for _, res := range s.searchResults {
		resultForPrint := fmt.Sprintf("Count for %s: %d", res.sourceName, res.numOfOccurrences)
		fmt.Println(resultForPrint)
	}
	fmt.Printf("Total: %d\n", s.sumOccurrences)
}

func (s *StrSearcher) addResult(input string, res int) {
	newResult := searchResult{
		sourceName:       input,
		numOfOccurrences: res,
	}
	s.mu.Lock()
	s.searchResults = append(s.searchResults, newResult)
	s.sumOccurrences += res
	s.mu.Unlock()
}

func (s *StrSearcher) isValidURL(input string) bool {
	u, err := url.Parse(input)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func (s *StrSearcher) isValidFile(input string) bool {
	_, err := os.Stat(input)
	return err == nil
}

func (s *StrSearcher) searchInFile(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			log.Printf("ERROR: %s: error in file closing: %s\n", filename, err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	var line string
	var count int
	for scanner.Scan() {
		line = scanner.Text()
		count += strings.Count(line, "Go")
	}

	if err = scanner.Err(); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *StrSearcher) searchHTTP(URL string) (int, error) {
	response, err := http.Get(URL)
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Printf("ERROR: %s: error in response body closing: %s\n", URL, err)
		}
	}(response.Body)

	scanner := bufio.NewScanner(response.Body)
	var line string
	count := 0
	for scanner.Scan() {
		line = scanner.Text()
		count += strings.Count(line, "Go")
	}
	return count, nil
}
