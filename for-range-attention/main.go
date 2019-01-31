package main

import "fmt"

func main()  {
	errDemo()
	rightDemo()
}

func errDemo() {
	slice := []int{0, 1, 2, 3}
	myMap := make(map[int]*int)

	for i, v := range slice {
		myMap[i] = &v
	}
	fmt.Println("=========errDemo=========")
	// 因为for-range会使用同一块内存去接受循环中的值
	// 所以上面的循环中，map[int]*int每一个Key指向的Value都是同一个地址
	prtMap(myMap)
}

func rightDemo() {
	slice := []int{0, 1, 2, 3}
	myMap := make(map[int]*int)

	for i, v := range slice {
		// 正确做法是将接收到的值复制到另外的内存块上
		num := v
		myMap[i] = &num
	}
	fmt.Println("=========rightDemo=========")
	prtMap(myMap)
}

func prtMap(m map[int]*int)  {
	for k, v := range m {
		fmt.Printf("map[%v]=%v\n", k, *v)
	}
}
