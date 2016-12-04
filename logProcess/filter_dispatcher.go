package main

import (
    "github.com/zieckey/goini"
    "github.com/guyannanfei25/flowprocess"
    "log"
    "fmt"
    "sync"
)

type FilterDispatcher struct {
    flowprocess.DefaultDispatcher
    badMid    map[string]int
    badLevel  int
    sync.Mutex
}

func (f *FilterDispatcher) Initialize(conf *goini.INI, section string) error {
    var ret bool
    if f.badLevel, ret = conf.SectionGetInt(section, "badLevel"); !ret {
        log.Printf("%s get badLevel err", f.GetName())
        return fmt.Errorf("%s get badLevel err", f.GetName())
    }

    f.badMid = make(map[string]int)

    return f.Init(conf, section)
}

func (f *FilterDispatcher) filter(item interface{}) {
    switch item := item.(type) {
    case *Context:
        f.Lock()
        defer f.Unlock()
        if item.level > f.badLevel {
            f.badMid[*(item.mid)] = 1
        }

        log.Printf("%s parse item[%#v] Success\n", f.GetName(), item)
    default:
        log.Printf("Get unknown type[%T]\n", item)
    }
}

func (f *FilterDispatcher) cleanUp() {
    for mid, _ := range f.badMid {
        log.Printf("%s filter %s is bad mid\n", f.GetName(), mid)
    }
}
