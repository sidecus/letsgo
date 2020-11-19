package exercise

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// MulticastDemo shows channel based multi cast
func MulticastDemo(maxWriterSleepMs int, maxReaderSleepMs int) {
	var wg sync.WaitGroup
	readerCnt, bufferLen := 3, 2

	fmt.Printf("%d readers, with buffer length: %d\n", readerCnt, bufferLen)

	mc, _ := CreateMultiCast(3, 2)

	// start readers
	for i, chr := range mc.chReaders {
		// reassign to per iteration scoped variable to avoid closure on loop variable
		id, ch := i, chr
		wg.Add(1)
		go func() {
			for v := range ch {
				fmt.Printf("Reader#%d processing %d\n", id, v)
				time.Sleep(time.Duration(rand.Intn(maxReaderSleepMs)) * time.Millisecond)
			}
			wg.Done()
		}()
	}

	// start writer
	wg.Add(1)
	go func() {
		for i := 0; i < 10; i++ {
			mc.chWriter <- i
			time.Sleep(time.Duration(rand.Intn(maxWriterSleepMs)) * time.Millisecond)
		}
		close(mc.chWriter)
		wg.Done()
	}()

	wg.Wait()
}
