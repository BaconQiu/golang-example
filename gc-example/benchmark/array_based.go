package main

import (
	"fmt"
	"time"
)

type (
	buffer  [windowSize]message
)

func (b *buffer) mkMessage(n int) message {
	m := make(message, 1024)
	for i := range m {
		m[i] = byte(n)
	}
	return m
}

func (b *buffer) pushMsg(highID int) {
	start := time.Now()
	m := b.mkMessage(highID)
	(*b)[highID%windowSize] = m
	elapsed := time.Since(start)
	if elapsed > worst {
		worst = elapsed
	}
}

func arrayBased() {
	worst = time.Duration(0)
	var b buffer
	for i := 0; i < msgCount; i++ {
		b.pushMsg(i)
	}
	fmt.Println("Worst push time based on array: ", worst)
}