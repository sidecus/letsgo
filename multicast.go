package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const bufLength = 2

// 1 v n multicast
func multicast(wg *sync.WaitGroup, producer func(chan<- int), consumer func(int, <-chan int), consumerCount int) {
	chIn := make(chan int)
	chOuts := make([]chan int, consumerCount)
	ret := make([]<-chan int, consumerCount)
	for i := range chOuts {
		chOuts[i] = make(chan int, bufLength)
		ret[i] = chOuts[i]
	}

	// start consumers
	for i, cho := range chOuts {
		// avoiding closure on loop variables
		id, ch := i, cho
		wg.Add(1)
		go func() {
			consumer(id, ch)
			wg.Done()
		}()
	}

	// start producer
	wg.Add(1)
	go func() {
		producer(chIn)
		close(chIn)
		wg.Done()
	}()

	// start multicasting
	wg.Add(1)
	go func() {
		for v := range chIn {
			for _, cho := range chOuts {
				select {
				case cho <- v: // success
				default: // full, get rid of oldest value and push new
					<-cho
					cho <- v
				}
			}
		}

		// close all output channels
		for _, cho := range chOuts {
			close(cho)
		}

		wg.Done()
	}()
}

func multicastDemo() {
	var wg sync.WaitGroup
	multicast(
		&wg,
		func(ch chan<- int) {
			for i := 0; i < 10; i++ {
				ch <- i
				time.Sleep(time.Duration(rand.Intn(1)) * time.Second)
			}
		},
		func(id int, ch <-chan int) {
			for v := range ch {
				fmt.Printf("Consumer#%d processing %d\n", id, v)
				time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
			}
		},
		2,
	)
	wg.Wait()
}
