package main

// #include <unistd.h>
import "C"
import (
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			//C.sleep(1)               // test 1
			time.Sleep(time.Second)  // test 2
			wg.Done()
		}()
	}

	wg.Wait()
	println("Done!")
	time.Sleep(time.Second * 3)
}


