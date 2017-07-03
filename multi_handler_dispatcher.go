package flowprocess

import (
    sj  "github.com/guyannanfei25/go-simplejson"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
    "log"
    "os"
    "errors"
)

var (
    ERRNONEEDPASSDOWN = errors.New("no need to pass down")
)

// 配置文件从ini改为json，json能够更好的处理复杂的层级关系
// http://www.guyannanfei25.site/2017/05/14/misc-talk/
type MultiHandlerDispatcher interface {
    Init(conf *sj.Json) error
    SetName(name string)
    GetName() string

    // 只有需要级联的dispatch才需要Register接口
    // 注册下游dispatcher
    DownRegister(d MultiHandlerDispatcher)

    // 注册上游dispatcher
    UpRegister(d MultiHandlerDispatcher)

    // 注册handler生成函数
    // Must before Start
    RegisterHandlerCreator(c HandlerCreator)

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

type Handler interface {
    // id for record which goroutine it stays
    Init(conf *sj.Json, id int) error
    Process(interface{}) (interface{}, error)
    Tick()
    Close()
}

type HandlerCreator func() Handler

type DefaultMultiHandlerDispatcher struct {
    name               string
    // 保存配置，以便用户需要获取相应参数
    Conf               *sj.Json
    // 改为map用于去重
    downDispatchers    map[MultiHandlerDispatcher]int
    upDispatchers      map[MultiHandlerDispatcher]int
    msgChan            chan interface{} // *item 
    msgMaxSize         int // 最大缓存大小
    concurrency        int // 并发执行数
    chanCount          uint32 // 当前协程数
    running            bool // 是否正在运行

    // handler生成函数
    creator            HandlerCreator
    handlers           []Handler

    wg                 sync.WaitGroup
    
    // 自定义logger
    logger             logger
    logLvl             LogLevel
    logGuard           sync.RWMutex
}

func (d *DefaultMultiHandlerDispatcher) Init(conf *sj.Json) error {
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

    d.name = conf.Get("name").MustString("DefaultMultiHandlerDispatcher")

    d.downDispatchers = make(map[MultiHandlerDispatcher]int)
    d.upDispatchers   = make(map[MultiHandlerDispatcher]int)
    d.handlers        = make([]Handler, d.concurrency)
    d.chanCount       = 0

    d.creator    = nil
    d.running    = false

    d.logger     =  log.New(os.Stderr, "", log.Flags())
    d.logLvl     =  LogLevelInfo

    return nil
}

func (d *DefaultMultiHandlerDispatcher) Start() {
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

func (d *DefaultMultiHandlerDispatcher) IsRunning() bool {
    return d.running
}

func (d *DefaultMultiHandlerDispatcher) SetName(name string) {
    d.name = name
}

func (d *DefaultMultiHandlerDispatcher) GetName() string {
    return d.name
}

func (d *DefaultMultiHandlerDispatcher) SetLogger(l logger, lvl LogLevel) {
    d.logGuard.Lock()
    defer d.logGuard.Unlock()

    d.logger = l
    d.logLvl = lvl
}

func (d *DefaultMultiHandlerDispatcher) getLogger() (logger, LogLevel) {
    d.logGuard.Lock()
    defer d.logGuard.Unlock()

    return d.logger, d.logLvl
}

func (d *DefaultMultiHandlerDispatcher) DownRegister(down MultiHandlerDispatcher) {
    d.downDispatchers[down] = 1
    down.UpRegister(d)
}

func (d *DefaultMultiHandlerDispatcher) UpRegister(up MultiHandlerDispatcher) {
    d.upDispatchers[up] = 1
}

// allow creator == nil, this MultiHandlerDispatcher use just for forward
func (d *DefaultMultiHandlerDispatcher) RegisterHandlerCreator (c HandlerCreator) {
    d.creator = c
}

func (d *DefaultMultiHandlerDispatcher) process(id int) {
    d.wg.Add(1)
    defer d.wg.Done()
    atomic.AddUint32(&d.chanCount, 1)

    var innerHandler Handler
    if d.creator != nil {
        // 每个工作协程一个handler对象，可以不用考虑多协程不安全问题
        innerHandler = d.creator()
        d.handlers[id] = innerHandler

        // init handler, if err then panic
        if err := innerHandler.Init(d.Conf, id); err != nil {
            d.logf(LogLevelFatal, "%s %dth handler init err[%s]\n", d.name, id, err)
        }
        defer innerHandler.Close()
    }

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

            if d.creator != nil {
                // handler handle item, if no need to pass down and no err happen just return ERRNONEEDPASSDOWN
                if ret, err = innerHandler.Process(item); err != nil {
                    if err != ERRNONEEDPASSDOWN {
                        d.logf(LogLevelError, "%s %dth process item[%v] err[%s]\n", 
                            d.name, id, item, err)
                    }
                    continue
                }
            } else {
                // just for forward
                ret = item
            }

            for sub, _ := range d.downDispatchers {
                sub.Dispatch(ret)
            }
        }
    }

    d.logf(LogLevelInfo, " %dth process quiting...\n", id)
    atomic.AddUint32(&d.chanCount, ^uint32(0))
}

func (d *DefaultMultiHandlerDispatcher) Dispatch(item interface{}) {
    d.msgChan <- item
}

func (d *DefaultMultiHandlerDispatcher) Ack(result interface{}) error {
    // TODO:是否合理???
    // DefaultMultiHandlerDispatcher不应该调用上游Dispatcher，应由继承者来实现
    // for up, _ := range d.upDispatchers {
        // up.Ack()
    // }
    return nil
}

// 一些定时任务，比如打印状态，更新缓存。。。
// 使用时建议使用一个单独协程，
// 由首个dispatcher调用，触犯后面dispatcher链
// framework := new(DefaultMultiHandlerDispatcher)
// go func(){
    // tick := time.NewTicker(time.Duration(20) * time.Second)
    // for {
        // <- tick.C
        // framework.Tick()
    // }
// }()
// defer tick.Stop()
func (d *DefaultMultiHandlerDispatcher) Tick() {
    // call every handler Tick
    if d.creator != nil {
        for _, h := range d.handlers {
            h.Tick()
        }
    }

    for sub, _ := range d.downDispatchers {
        sub.Tick()
    }
}

func (d *DefaultMultiHandlerDispatcher) logf(lvl LogLevel, line string, args ...interface{}) {
    logger, logLvl := d.getLogger()

    if logger == nil {
        return
    }

    if logLvl > lvl {
        return
    }

    logger.Output(2, fmt.Sprintf("[%-7s %s] %s", lvl, d.GetName(), fmt.Sprintf(line, args...)))
    if logLvl == LogLevelFatal {
        os.Exit(1)
    }
}

func (d *DefaultMultiHandlerDispatcher) Close() {
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
