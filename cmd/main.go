package main

import (
	"bufio"
	"log"
	"os"
	"sync"
	"vk-task/searcher"
)

func main() {
	var goSearcher searcher.StringSearcher
	wg := &sync.WaitGroup{}
	goSearcher = searcher.NewStrSearcher("Go", 5)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		wg.Add(1)
		inputStr := scanner.Text()
		go func() {
			defer wg.Done()
			err := goSearcher.MakeSearch(inputStr)
			if err != nil {
				log.Printf("ERROR: error for input %s: %s\n", inputStr, err)
			}
		}()
	}
	wg.Wait()
	goSearcher.PrintSearchResults()
}
