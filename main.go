package main

import (
	"fmt"
	"runtime"
	"strconv"
)

type workType struct {
	name  string
	start func()
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
	{"MultiCast - Eager writer", func() { multicastDemo(50, 200) }},
	{"MultiCast - Lazy writer", func() { multicastDemo(200, 50) }},
	{"RaftDemo", raftDemo},
}

func selectWork() int {
	start := 0
	end := len(workItems) - 1
	selection := -1
	for selection < start || selection > end {
		fmt.Printf("Select exercise to run (%d - %d):", 0, len(workItems)-1)
		var str string
		fmt.Scanln(&str)
		sel, err := strconv.Atoi(str)
		if err != nil {
			sel = -1
		}
		selection = sel
	}

	return selection
}

func main() {
	for i, v := range workItems {
		fmt.Printf("%d: %s\n", i, v.name)
	}

	for {
		selection := selectWork()
		workItem := workItems[selection]
		fmt.Printf("You selected %s, starting work\n", workItem.name)
		workItem.start()
	}
}
