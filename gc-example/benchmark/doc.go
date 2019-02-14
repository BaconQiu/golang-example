package main

/**

该测试将不断地向一个限定缓冲区大小的buffer中推送消息，
旧的消息将会不断地过期并成为垃圾需要进行回收，这要求内存堆需要一直保持较大的状态，这很重要，因为在回收的阶段整个内存堆都需要进行扫描以确定是否有内存引用。
这也是为什么GC的运行时间和存活的内存对象和指针数目成正比例关系的原因。

参考：
https://juejin.im/post/5c62d45ee51d457fa44f4404
https://making.pusher.com/golangs-real-time-gc-in-theory-and-practice/

 */
