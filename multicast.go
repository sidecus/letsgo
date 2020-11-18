package main

import (
	"errors"
	"sync"
)

// cannot use const... wth, how can I define error constants (or at least readonly variables) in go?
var errorOutputBufLen = errors.New("outputBufLen is invalid")
var errorConsumerCount = errors.New("consumerCount is invalid")
var errorInvalidCallBack = errors.New("consumer or producer callback is invalid")
var errorUsedMultiCast = errors.New("multicast already used, cannot start again")

// MultiCast struct
type MultiCast struct {
	chIn         chan int
	chOuts       []chan int
	outputBufLen int
	consumerCnt  int
	wg           sync.WaitGroup
	used         bool
}

// NewMultiCast initializes a new NewMultiCast object
func NewMultiCast(outputBufLen int, consumerCount int) (*MultiCast, error) {
	if outputBufLen <= 0 {
		return nil, errorOutputBufLen
	}
	if consumerCount <= 0 {
		return nil, errorConsumerCount
	}
	chIn := make(chan int)
	chOuts := make([]chan int, consumerCount)
	for i := range chOuts {
		chOuts[i] = make(chan int, outputBufLen)
	}

	return &MultiCast{
		chIn:         chIn,
		chOuts:       chOuts,
		outputBufLen: outputBufLen,
		consumerCnt:  consumerCount,
		used:         false,
	}, nil
}

// Run starts the multi casting logic for the given producer and consumer
func (mc *MultiCast) Run(producer func(chan<- int), consumer func(int, <-chan int)) error {
	if mc.used {
		return errorUsedMultiCast
	}

	if producer == nil || consumer == nil {
		return errorInvalidCallBack
	}

	wg := &mc.wg

	// start consumers
	for i, cho := range mc.chOuts {
		// avoiding closure on loop variables
		id, ch := i, cho
		wg.Add(1)
		go mc.consume(consumer, id, ch)
	}

	// start producer
	wg.Add(1)
	go mc.produce(producer)

	// start casting
	wg.Add(1)
	go mc.cast()

	// if we reach here we are done
	return nil
}

func (mc *MultiCast) produce(producer func(chan<- int)) {
	producer(mc.chIn)

	// close channel and signal done
	close(mc.chIn)
	mc.wg.Done()
}

func (mc *MultiCast) consume(consumer func(int, <-chan int), id int, ch chan int) {
	consumer(id, ch)
	// signal done
	mc.wg.Done()
}

func (mc *MultiCast) cast() {
	for v := range mc.chIn {
		for _, cho := range mc.chOuts {
			select {
			case cho <- v: // success
			default: // full, get rid of oldest value and push new
				<-cho
				cho <- v
			}
		}
	}

	// close all output channels
	for _, cho := range mc.chOuts {
		close(cho)
	}

	mc.wg.Done()
}
