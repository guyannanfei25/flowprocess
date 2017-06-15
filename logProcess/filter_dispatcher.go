package main

import (
    sj  "github.com/bitly/go-simplejson"
    "github.com/guyannanfei25/flowprocess"
    "log"
    "sync"
)

type FilterDispatcher struct {
    flowprocess.DefaultDispatcher
    badMid    map[string]int
    badLevel  int
    sync.Mutex
}

func (f *FilterDispatcher) Init(conf *sj.Json) error {
    f.badLevel = conf.Get("badLevel").MustInt(50)

    f.badMid = make(map[string]int)

    return f.DefaultDispatcher.Init(conf)
}

func (f *FilterDispatcher) filter(item interface{}) (interface{}, error) {
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
    return item, nil
}

func (f *FilterDispatcher) cleanUp() {
    for mid, _ := range f.badMid {
        log.Printf("%s filter %s is bad mid\n", f.GetName(), mid)
    }
}

func (f *FilterDispatcher) Close() {
    f.DefaultDispatcher.Close()
    f.cleanUp()
}
