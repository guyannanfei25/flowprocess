package main

import (
    sj  "github.com/bitly/go-simplejson"
    "log"
    "sync"
)

type FilterDispatcher struct {
    badMid    map[string]int
    badLevel  int
    id        int
    sync.Mutex
}

func (f *FilterDispatcher) Init(conf *sj.Json, id int) error {
    f.badLevel = conf.Get("badLevel").MustInt(50)

    f.badMid = make(map[string]int)
    f.id = id
    log.Printf("%dth FilterDispatcher init\n", id)

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

func (f *FilterDispatcher) Tick() {
}

func (f *FilterDispatcher) Close() {
    f.cleanUp()
    log.Printf("%dth FilterDispatcher close\n", f.id)
}
