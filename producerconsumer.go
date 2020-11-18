package main

import (
	"fmt"
	"math/rand"
	"time"
)

func producer(id int, ch chan string) string {
	for {
		i := rand.Intn(100)
		ch <- fmt.Sprintf("%d=>%d", id, i)

		// sleep i/10 seconds
		time.Sleep(time.Duration(i/10) * time.Second)
	}
}

// ProdConMain demos producer consumer scenario using channels
func ProdConMain() {
	pn := 5
	ch := make(chan string)
	for i := 0; i < pn; i++ {
		go producer(i, ch)
	}

	for {
		p := <-ch
		fmt.Printf("Consumed %s\n", p)
	}
}
