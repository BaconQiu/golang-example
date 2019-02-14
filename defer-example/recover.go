package main

import (
	"fmt"
	"os"
	"sync"
)

func recoverDemo() {
	defer fmt.Println("defer main")
	var user = os.Getenv("USER_")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		defer func() {
			fmt.Println("defer caller")
			if err := recover(); err != nil {
				fmt.Println("recover success. err: ", err)
			}
		}()

		func() {
			defer func() {
				fmt.Println("defer here")
			}()

			if user == "" {
				panic("should set user env.")
			}

			// 此处不会执行
			fmt.Println("after panic")
		}()

	}()
	wg.Wait()

	fmt.Println("end of main function")
}