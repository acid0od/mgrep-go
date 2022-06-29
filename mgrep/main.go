package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"io/ioutil"
	"mgrep/worker"
	"mgrep/worklist"
	"path/filepath"
	"sync"
)

func discoverDirs(wl *worklist.Worklist, path string) {
	entries, err := ioutil.ReadDir(path)

	if err != nil {
		fmt.Printf("Error reading directory: %v", err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			nextPath := filepath.Join(path, entry.Name())
			discoverDirs(wl, nextPath)
		} else {
			wl.Add(worklist.NewJob(filepath.Join(path, entry.Name())))
		}
	}
}

var args struct {
	SearchTerm string `arg:"positional,required"`
	SearchDir  string `arg:"positional"`
}

func main() {
	arg.MustParse(&args)

	var workersWg sync.WaitGroup
	fmt.Printf("Searching for %s in %s\n", args.SearchTerm, args.SearchDir)
	wl := worklist.New(100)
	results := make(chan worker.Result, 100)
	numberWorkers := 10
	workersWg.Add(1)

	go func() {
		defer workersWg.Done()
		discoverDirs(&wl, args.SearchDir)
		wl.Finalize(numberWorkers)
	}()

	for i := 0; i < numberWorkers; i++ {
		workersWg.Add(1)
		go func() {
			defer workersWg.Done()
			for {
				workEntry := wl.Next()
				if workEntry.Path == "" {
					return
				} else {
					workerResult := worker.FindInFile(workEntry.Path, args.SearchTerm)
					if workerResult != nil {
						for _, r := range workerResult.Inner {
							results <- r
						}
					}
				}
			}
		}()
	}

	blockWorkerWg := make(chan struct{})
	go func() {
		workersWg.Wait()
		close(blockWorkerWg)
	}()

	var displayWg sync.WaitGroup

	displayWg.Add(1)

	go func() {
		for {
			select {
			case r := <-results:
				fmt.Println(r.Print())
			case <-blockWorkerWg:
				if len(results) == 0 {
					displayWg.Done()
					return
				}
			}
		}
	}()

	displayWg.Wait()
}
