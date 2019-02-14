package main

import (
	"fmt"
	"time"
)

type (
	bufferMap  map[int]message
)

func (b *bufferMap) mkMessage(n int) message {
	m := make(message, 1024)
	for i := range m {
		m[i] = byte(n)
	}
	return m
}

func (b *bufferMap) pushMsg(highID int) {
	start := time.Now()
	m := b.mkMessage(highID)
	(*b)[highID%windowSize] = m
	elapsed := time.Since(start)
	if elapsed > worst {
		worst = elapsed
	}
}

func mapBased() {
	worst = time.Duration(0)
	var b bufferMap = make(map[int]message, windowSize)
	for i := 0; i < msgCount; i++ {
		b.pushMsg(i)
	}
	fmt.Println("Worst push time based on map: ", worst)
}