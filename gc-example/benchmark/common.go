package main

import "time"

const (
	windowSize = 200000
	msgCount   = 1000000
)

type (
	message []byte
)

var worst time.Duration