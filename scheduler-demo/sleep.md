
该程序用于展示在GolangGPM模型中，G（Goroutine）遇到系统调用阻塞和非系统调用阻塞的时候，调度模型将如何处理。

该程序示例取自 http://wudaijun.com/2018/01/go-scheduler/ , 关于调度模型的详细描述可参考该链接。这里不做任何介绍。

test1结果(调用C.sleep(1))
````
$ GODEBUG=schedtrace=1000 ./sleep
SCHED 0ms: gomaxprocs=2 idleprocs=0 threads=6 spinningthreads=1 idlethreads=1 runqueue=0 [0 135]
SCHED 1001ms: gomaxprocs=2 idleprocs=2 threads=1004 spinningthreads=0 idlethreads=3 runqueue=0 [0 0]
Done!
SCHED 2010ms: gomaxprocs=2 idleprocs=2 threads=1004 spinningthreads=0 idlethreads=1000 runqueue=0 [0 0]
SCHED 3011ms: gomaxprocs=2 idleprocs=2 threads=1004 spinningthreads=0 idlethreads=1000 runqueue=0 [0 0]
SCHED 4018ms: gomaxprocs=2 idleprocs=2 threads=1004 spinningthreads=0 idlethreads=1000 runqueue=0 [0 0]
````

test1结果(time.Sleep(time.Second))
````
$ GODEBUG=schedtrace=1000 ./sleep
SCHED 0ms: gomaxprocs=2 idleprocs=0 threads=6 spinningthreads=0 idlethreads=1 runqueue=0 [0 0]
Done!
SCHED 1005ms: gomaxprocs=2 idleprocs=2 threads=7 spinningthreads=0 idlethreads=3 runqueue=0 [0 0]
SCHED 2015ms: gomaxprocs=2 idleprocs=2 threads=7 spinningthreads=0 idlethreads=3 runqueue=0 [0 0]
SCHED 3025ms: gomaxprocs=2 idleprocs=2 threads=7 spinningthreads=0 idlethreads=3 runqueue=0 [0 0]
````

下面介绍schedtrace日志每一行的字段意义(摘自上面链接)

- SCHED：调试信息输出标志字符串，代表本行是goroutine scheduler的输出；
- 1001ms：即从程序启动到输出这行日志的时间；
- gomaxprocs: P的数量；
- idleprocs: 处于idle状态的P的数量；通过gomaxprocs和idleprocs的差值，我们就可知道执行go代码的P的数量；
- threads: os threads的数量，包含scheduler使用的m数量，加上runtime自用的类似sysmon这样的thread的数量；
- spinningthreads: 处于自旋状态的os thread数量；
- idlethread: 处于idle状态的os thread的数量；
- runqueue： go scheduler全局队列中G的数量；
- [0 0]: 分别为2个P的local queue中的G的数量。

可以看出，time.Sleep并没有使用系统调用，使得仅仅只有G阻塞了，M不会阻塞。但在使用cgo sleep的情况下，可以看到大量的闲置M。



