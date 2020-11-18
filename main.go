package main

import (
	"fmt"
	"runtime"
	"strconv"
)

type workType struct {
	name  string
	entry func()
}

var workItems = []workType{
	{"Crawler", CrawlerMain},
	{"BSTComparison", BSTComparisonMain},
	{"Web Server", WebStartUp},
	{"CPUNumber", func() {
		fmt.Println(runtime.NumCPU())
	}},
	{"Defer", func() {
		defer fmt.Println("Hello")
		defer fmt.Println("!")
		fmt.Println("World")
	}},
	{"ProducerConsumer", ProdConMain},
	{"multicastDemo", multicastDemo},
}

func selectWork() (int, error) {
	start := 0
	end := len(workItems) - 1
	selection := -1
	for selection < start || selection > end {
		fmt.Printf("Select exercise to run (%d - %d):", 0, len(workItems)-1)
		var str string
		fmt.Scanln(&str)
		sel, err := strconv.Atoi(str)
		if err != nil {
			selection = -1
		} else {
			selection = sel
		}
	}

	return selection, nil
}

func main() {
	for i, v := range workItems {
		fmt.Printf("%d: %s\n", i, v.name)
	}

	selection, err := selectWork()
	if err != nil {
		return
	}

	workItem := workItems[selection]
	fmt.Printf("You selected %s, starting work\n", workItem.name)
	workItem.entry()
}
