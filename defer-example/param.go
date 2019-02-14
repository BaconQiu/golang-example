package main

import "fmt"

// 被推迟函数的实参在推迟执行的时候就会被求值，而不是在调用执行的时候才求值

func trace(s string) string {
	fmt.Println("entering: ", s)
	return s
}

func un(s string) {
	fmt.Println("leaving: ", s)
}

func foo() {
	defer un(trace("foo"))
	fmt.Println("in foo")
}

func bar() {
	defer un(trace("bar"))
	fmt.Println("in bar")
	foo()
}

