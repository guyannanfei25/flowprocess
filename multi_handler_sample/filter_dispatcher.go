package main

import (
    sj  "github.com/bitly/go-simplejson"
    "log"
    "sync"
)

type FilterDispatcher struct {
    badMid    map[string]int
    badLevel  int
    sync.Mutex
}

func (f *FilterDispatcher) Init(conf *sj.Json) error {
    f.badLevel = conf.Get("badLevel").MustInt(50)

    f.badMid = make(map[string]int)

    return nil
}

func (f *FilterDispatcher) Process(item interface{}) (interface{}, error) {
    switch item := item.(type) {
    case *Context:
        f.Lock()
        defer f.Unlock()
        if item.level > f.badLevel {
            f.badMid[*(item.mid)] = 1
        }

        log.Printf("parse item[%#v] Success\n", item)
    default:
        log.Printf("Get unknown type[%T]\n", item)
    }
    return item, nil
}

func (f *FilterDispatcher) cleanUp() {
    for mid, _ := range f.badMid {
        log.Printf("filter %s is bad mid\n", mid)
    }
}

func (f *FilterDispatcher) Close() {
    f.cleanUp()
}
