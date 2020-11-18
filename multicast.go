package main

import (
	"errors"
)

// cannot use const and there is no readonly... wth, how can I define these better in go?
var errorReaderCount = errors.New("reader count is invalid")
var errorReaderChannelBufferLength = errors.New("reader buffer length is invalid")

// MultiCast struct
type MultiCast struct {
	chWriter     chan int
	chReaders    []chan int
	readerCount  int
	readerBufLen int
}

// CreateMultiCast starts a UDP like multi cast, no reliability gaurantee for slow readers
func CreateMultiCast(readerCount int, readerBufLen int) (*MultiCast, error) {
	if readerCount <= 0 {
		return nil, errorReaderCount
	}
	if readerBufLen <= 0 {
		return nil, errorReaderChannelBufferLength
	}

	// create writer and reader channels
	chWriter := make(chan int)
	chReaders := make([]chan int, readerCount)
	for i := range chReaders {
		chReaders[i] = make(chan int, readerBufLen)
	}

	mc := &MultiCast{
		chWriter:     chWriter,
		chReaders:    chReaders,
		readerBufLen: readerBufLen,
		readerCount:  readerCount,
	}

	// start multicasting goroutine
	go mc.cast()

	return mc, nil
}

// cast starts multicasting - read from writer channel and fan out to all reader channels
func (mc *MultiCast) cast() {
	for v := range mc.chWriter {
		// cast to each reader channel (lossy for now, like UDP)
		for _, chr := range mc.chReaders {
			select {
			case chr <- v: // success
			default:
				// buffer full, get rid of oldest value and push new
				<-chr
				chr <- v
			}
		}
	}

	// input chanel closed, now we need to close all output channels
	for _, chr := range mc.chReaders {
		close(chr)
	}
}
