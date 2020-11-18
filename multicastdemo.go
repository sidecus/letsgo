package main

import (
	"fmt"
	"math/rand"
	"time"
)

func multicastDemo(maxProducerSleepMs int, maxConsumerSleepMs int) {
	mc, _ := NewMultiCast(2, 3)
	mc.Run(
		func(ch chan<- int) {
			for i := 0; i < 10; i++ {
				ch <- i
				time.Sleep(time.Duration(rand.Intn(maxProducerSleepMs)) * time.Millisecond)
			}
		},
		func(id int, ch <-chan int) {
			for v := range ch {
				fmt.Printf("Consumer#%d processing %d\n", id, v)
				time.Sleep(time.Duration(rand.Intn(maxConsumerSleepMs)) * time.Millisecond)
			}
		},
	)
	mc.wg.Wait()
}
