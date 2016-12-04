package flowprocess

import (
    "github.com/zieckey/goini"
    "fmt"
    "sync/atomic"
    "time"
)

type Dispatcher interface {
    Init(conf *goini.INI, section string) error
    SetName(name string)
    GetName() string

    // 只有需要级联的dispatch才需要Register接口
    // Register(d *Dispatcher)

    // 注册下游dispatcher
    DownRegister(d Dispatcher)

    // 注册上游dispatcher
    UpRegister(d Dispatcher)

    // Dispatch前预处理函数
    SetPreFunc(preFunctor Functor)

    // Dispatch后处理函数
    SetSufFuc(sufFunctor Functor)

    // 开始运行，在注册好dispatcher并设置好预处理函数之后
    Start()

    // 是否正在运行
    IsRunning() bool

    Dispatch(item interface{})
    Ack(result interface{}) error
    Close()
}

// 函数对象
type Functor func(interface{})

// type Item struct {
// }
// 仅做中转，由具体的用户定制的dispatcher去转换
// 自定义类型外部需要导入
// type Item interface{}

// 返回结果，用户定制，比如统计成功率
// type Result interface{}

type DefaultDispatcher struct {
    name               string
    // 改为map用于去重
    downDispatchers    map[Dispatcher]int
    upDispatchers      map[Dispatcher]int
    msgChan            chan interface{} // *item 
    msgMaxSize         int // 最大缓存大小
    concurrency        int // 并发执行数
    chanCount          uint32 // 当前协程数
    running            bool // 是否正在运行

    // 预处理函数
    preFunctor         Functor

    // 最后处理函数
    sufFunctor         Functor

    // Close之前清理函数
    cleanUp            func()
}

func (d *DefaultDispatcher) Init(conf *goini.INI, section string) error {
    var ret bool
    if d.concurrency, ret = conf.SectionGetInt(section, "concurrency"); !ret {
        fmt.Println("DefaultDispatcher Get concurrency conf err")
        return fmt.Errorf("DefaultDispatcher Get concurrency conf err")
    }

    if d.concurrency < 1 {
        d.concurrency = 1
    }

    if d.msgMaxSize, ret = conf.SectionGetInt(section, "msgMaxSize"); !ret {
        fmt.Println("DefaultDispatcher Get msgMaxSize conf err")
        return fmt.Errorf("DefaultDispatcher Get msgMaxSize conf err")
    }

    if d.msgMaxSize < 0 {
        d.msgMaxSize = 0 // 无缓存
    }
    d.msgChan = make(chan interface{},  d.msgMaxSize)

    if d.name, ret = conf.SectionGet(section, "name"); !ret {
        fmt.Println("DefaultDispatcher Get name conf err")
        return fmt.Errorf("DefaultDispatcher Get name conf err")
    }

    d.downDispatchers = make(map[Dispatcher]int)
    d.upDispatchers = make(map[Dispatcher]int)
    d.chanCount     = 0

    d.preFunctor = nil
    d.sufFunctor = nil
    d.cleanUp    = nil
    d.running    = false

    return nil
}

func (d *DefaultDispatcher) Start() {
    if d.running {
        fmt.Printf("%s is already running!\n", d.name)
        return
    }

    for i := 0; i < d.concurrency; i++ {
        go d.process(i)
    }

    // Make sure down and up dispatcher running
    // TODO:是否由一个统一启动,还是自动启动自己的？？
    for down, _ := range d.downDispatchers {
        down.Start()
        
        // to ensure downregister start success
        for {
            if !down.IsRunning() {
                time.Sleep(time.Second)
            } else {
                break
            }
        }
    }

    // 只管下游
    // for up, _ := range d.upDispatchers {
        // up.Start()
    // }

    d.running = true
    fmt.Printf("%s start running!\n", d.name)
}

func (d *DefaultDispatcher) IsRunning() bool {
    return d.running
}

func (d *DefaultDispatcher) SetName(name string) {
    d.name = name
}

func (d *DefaultDispatcher) GetName() string {
    return d.name
}

func (d *DefaultDispatcher) DownRegister(down Dispatcher) {
    // d.downDispatchers = append(d.downDispatchers, down)
    d.downDispatchers[down] = 1
    down.UpRegister(d)
}

func (d *DefaultDispatcher) UpRegister(up Dispatcher) {
    // d.upDispatchers = append(d.upDispatchers, up)
    d.upDispatchers[up] = 1
}

func (d *DefaultDispatcher) SetPreFunc (preFunctor Functor) {
    d.preFunctor = preFunctor
}

func (d *DefaultDispatcher) SetSufFuc (sufFunctor Functor) {
    d.sufFunctor = sufFunctor
}

func (d *DefaultDispatcher) SetCleanUp (cleanUp func()) {
    d.cleanUp = cleanUp
}

func (d *DefaultDispatcher) process(id int) {
    atomic.AddUint32(&d.chanCount, 1)
    fmt.Printf("%s %d th process starting...\n", d.name, id)

PROCESS_MAIN:
    for {
        select {
        case item, ok := <- d.msgChan:
            if !ok {
                break PROCESS_MAIN
            }

            // dispatch之前预处理函数
            if d.preFunctor != nil {
                d.preFunctor(item)
            }

            for sub, _ := range d.downDispatchers {
                sub.Dispatch(item)
            }

            // dispatch之后处理函数
            if d.sufFunctor != nil {
                d.sufFunctor(item)
            }
        }
    }

    fmt.Printf("%s %d th process quiting...\n", d.name, id)
    atomic.AddUint32(&d.chanCount, ^uint32(0))
}

// func (d *DefaultDispatcher) Dispatch(item *Item) {
func (d *DefaultDispatcher) Dispatch(item interface{}) {
    d.msgChan <- item
}

// func (d *DefaultDispatcher) Ack(result Result) error {
func (d *DefaultDispatcher) Ack(result interface{}) error {
    // TODO:是否合理???
    // DefaultDispatcher不应该调用上游Dispatcher，应由继承者来实现
    // for up, _ := range d.upDispatchers {
        // up.Ack()
    // }
    return nil
}

func (d *DefaultDispatcher) Close() {
    // TODO:是否需要关闭其他的
    close(d.msgChan)

    for {
        if d.chanCount != 0 {
            // 等待清理现场
            time.Sleep(time.Second)
        } else {
            break
        }
    }
    d.running = false

    // 首先关闭协程，因为清理工作可能会清理掉协程需要的资源
    if d.cleanUp != nil {
        d.cleanUp()
    }

    // 关闭下游dispatcher
    for sub, _ := range d.downDispatchers {
        if sub.IsRunning() {
            sub.Close()
        }

        for {
            if sub.IsRunning() {
                time.Sleep(time.Second)
            } else {
                break
            }
        }
    }

    fmt.Printf("%s exit Success!\n", d.name)
}
