
package main

import (
	"log"
	"sync"
)

// 单个生产者，多个消费者
func spmc() {
	log.SetFlags(0)

	// ...
	const MaxNumber = 100
	const NumReceivers = 10

	wgReceivers := sync.WaitGroup{}
	wgReceivers.Add(NumReceivers)

	// ...
	dataCh := make(chan int, 100)

	var value = 0

	// the sender
	go func() {
		for {
			if value >= MaxNumber {
				// the only sender can close the channel safely.
				log.Println("attempt to close dataCh")
				close(dataCh)
				return
			} else {
				dataCh <- value
				value ++
			}
		}
	}()

	// receivers
	for i := 0; i < NumReceivers; i++ {
		go func(numReceiver int) {
			defer wgReceivers.Done()

			// receive values until dataCh is closed and
			// the value buffer queue of dataCh is empty.
			for value := range dataCh {
				log.Printf("%d receive val: %d", numReceiver, value)
			}
		}(i)
	}

	wgReceivers.Wait()
}
