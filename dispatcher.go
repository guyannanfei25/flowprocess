package flowprocess

import (
    sj  "github.com/bitly/go-simplejson"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
    "log"
    "os"
)

// 配置文件从ini改为json，json能够更好的处理复杂的层级关系
// http://www.guyannanfei25.site/2017/05/14/misc-talk/
type Dispatcher interface {
    Init(conf *sj.Json) error
    SetName(name string)
    GetName() string

    // 只有需要级联的dispatch才需要Register接口
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

    // 自定义logger
    SetLogger(l logger, lvl LogLevel)
    getLogger() (logger, LogLevel)

    // 一些定时任务，比如打印状态，更新缓存。。。
    Tick()
    Close()
}

// 函数对象
// 针对prefunc 返回错误则会终止后续的处理
// 可以定制返回结果，后续操作继续操作
type Functor func(interface{}) (interface{}, error)

// type Item struct {
// }
// 仅做中转，由具体的用户定制的dispatcher去转换
// 自定义类型外部需要导入
// type Item interface{}

// 返回结果，用户定制，比如统计成功率
// type Result interface{}

type DefaultDispatcher struct {
    name               string
    // 保存配置，以便用户需要获取相应参数
    Conf               *sj.Json
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
    wg                 sync.WaitGroup
    
    // 自定义logger
    logger             logger
    logLvl             LogLevel
    logGuard           sync.RWMutex
}

func (d *DefaultDispatcher) Init(conf *sj.Json) error {
    d.Conf = conf
    d.concurrency = conf.Get("concurrency").MustInt(1)
    if d.concurrency < 1 {
        d.concurrency = 1
    }

    d.msgMaxSize = conf.Get("msgMaxSize").MustInt(0)

    if d.msgMaxSize < 0 {
        d.msgMaxSize = 0 // 无缓存
    }
    d.msgChan = make(chan interface{},  d.msgMaxSize)

    d.name = conf.Get("name").MustString("DefaultName")

    d.downDispatchers = make(map[Dispatcher]int)
    d.upDispatchers = make(map[Dispatcher]int)
    d.chanCount     = 0

    d.preFunctor = nil
    d.sufFunctor = nil
    d.running    = false

    d.logger     =  log.New(os.Stderr, "", log.Flags())
    d.logLvl     =  LogLevelInfo

    return nil
}

func (d *DefaultDispatcher) Start() {
    if d.running {
        d.logf(LogLevelInfo, " is already running!\n")
        return
    }

    for i := 0; i < d.concurrency; i++ {
        go d.process(i)
    }

    // Make sure down and up dispatcher running
    // 由上一级 dispatcher 启动所有下一级
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

    d.running = true
    d.logf(LogLevelInfo, " start running!\n")
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

func (d *DefaultDispatcher) SetLogger(l logger, lvl LogLevel) {
    d.logGuard.Lock()
    defer d.logGuard.Unlock()

    d.logger = l
    d.logLvl = lvl
}

func (d *DefaultDispatcher) getLogger() (logger, LogLevel) {
    d.logGuard.Lock()
    defer d.logGuard.Unlock()

    return d.logger, d.logLvl
}

func (d *DefaultDispatcher) DownRegister(down Dispatcher) {
    d.downDispatchers[down] = 1
    down.UpRegister(d)
}

func (d *DefaultDispatcher) UpRegister(up Dispatcher) {
    d.upDispatchers[up] = 1
}

func (d *DefaultDispatcher) SetPreFunc (preFunctor Functor) {
    d.preFunctor = preFunctor
}

func (d *DefaultDispatcher) SetSufFuc (sufFunctor Functor) {
    d.sufFunctor = sufFunctor
}

func (d *DefaultDispatcher) process(id int) {
    d.wg.Add(1)
    defer d.wg.Done()
    atomic.AddUint32(&d.chanCount, 1)
    d.logf(LogLevelInfo, " %dth process starting...\n", id)

PROCESS_MAIN:
    for {
        select {
        case item, ok := <- d.msgChan:
            if !ok {
                break PROCESS_MAIN
            }

            // 经过用户业务处理之后的返回值，可以简简单单返回item自身
            var ret interface{}
            var err error

            // dispatch之前预处理函数
            // 当需要中断处理，即不需要后续的dispatcher处理则需要返回错误
            if d.preFunctor != nil {
                if ret, err = d.preFunctor(item); err != nil {
                    continue
                }
            } else {
                // 如果没有设置用户自己业务，则仅仅做转发使用
                ret = item
            }

            for sub, _ := range d.downDispatchers {
                sub.Dispatch(ret)
            }

            // dispatch之后处理函数
            if d.sufFunctor != nil {
                d.sufFunctor(item)
            }
        }
    }

    d.logf(LogLevelInfo, " %dth process quiting...\n", id)
    atomic.AddUint32(&d.chanCount, ^uint32(0))
}

func (d *DefaultDispatcher) Dispatch(item interface{}) {
    d.msgChan <- item
}

func (d *DefaultDispatcher) Ack(result interface{}) error {
    // TODO:是否合理???
    // DefaultDispatcher不应该调用上游Dispatcher，应由继承者来实现
    // for up, _ := range d.upDispatchers {
        // up.Ack()
    // }
    return nil
}

// 一些定时任务，比如打印状态，更新缓存。。。
// 使用时建议使用一个单独协程，
// 由首个dispatcher调用，触犯后面dispatcher链
// framework := new(DefaultDispatcher)
// go func(){
    // tick := time.NewTicker(time.Duration(20) * time.Second)
    // for {
        // <- tick.C
        // framework.Tick()
    // }
// }()
// defer tick.Stop()
func (d *DefaultDispatcher) Tick() {
    for sub, _ := range d.downDispatchers {
        sub.Tick()
    }
}

func (d *DefaultDispatcher) logf(lvl LogLevel, line string, args ...interface{}) {
    logger, logLvl := d.getLogger()

    if logger == nil {
        return
    }

    if logLvl > lvl {
        return
    }

    logger.Output(2, fmt.Sprintf("[%-7s %s] %s", lvl, d.GetName(), fmt.Sprintf(line, args...)))
}

func (d *DefaultDispatcher) Close() {
    close(d.msgChan)
    d.wg.Wait()

    d.running = false

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

    d.logf(LogLevelInfo, " exit Success\n")
}
