package main


/**
return XXX
经过编译之后会变成三条指令，如下：
1. 返回值=XXX
2. 调用defer函数
3. 空的return

所以以下f()拆解的结果为：

func f() (r int) {
     t := 5

     // 1. 赋值指令
     r = t

     // 2. defer被插入到赋值与返回之间执行，这个例子中返回值r没被修改过
     func() {
         t = t + 5
     }

     // 3. 空的return指令
     return
}

返回的结果为5
 */

func f() (r int) {
	t := 5

	defer func() {
		t = t + 5
	}()

	return t
}
