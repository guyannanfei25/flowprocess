package main

import (
    "github.com/zieckey/goini"
    "github.com/guyannanfei25/flowprocess"
    "log"
    "sync"
)

type PvDispatcher struct {
    flowprocess.DefaultDispatcher
    pvMap    map[string]int
    sync.Mutex
}

// 自定义初始化可以使用其他名称，然后自己内部调用父类初始化方法
// func (p *PvDispatcher) Initialize(conf *goini.INI, section string) error {
    // // 各自dispatcher需要自己初始化的地方
    // p.pvMap = make(map[string]int)
    // return p.Init(conf, section)
// }

// 或者覆盖父类初始化方法，内部调用
func (p *PvDispatcher) Init(conf *goini.INI, section string) error {
    p.pvMap = make(map[string]int)
    return p.DefaultDispatcher.Init(conf, section)
}

func (p *PvDispatcher) pvParse(item interface{}) (interface{}, error)  {
    switch item := item.(type) {
    case *Context:
        p.Lock()
        defer p.Unlock()
        if _, ok := p.pvMap[*(item.mid)]; !ok {
            p.pvMap[*(item.mid)] = 0
        }

        p.pvMap[*(item.mid)] += 1

        log.Printf("%s parse item[%#v] Success\n", p.GetName(), item)
    default:
        log.Printf("Get unknown type[%T]\n", item)
    }

    return item, nil
}

func (p *PvDispatcher) cleanUp() {
    for mid, pv := range p.pvMap {
        log.Printf("%s pv is %d\n", mid, pv)
    }
}

func (p *PvDispatcher) Close() {
    p.DefaultDispatcher.Close()
    p.cleanUp()
}
