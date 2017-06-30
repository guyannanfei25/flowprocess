package main

import (
    sj  "github.com/bitly/go-simplejson"
    "log"
    "sync"
)

type PvDispatcher struct {
    pvMap    map[string]int
    sync.Mutex
}

func (p *PvDispatcher) Init(conf *sj.Json) error {
    p.pvMap = make(map[string]int)
    return nil
}

func (p *PvDispatcher) Process(item interface{}) (interface{}, error)  {
    switch item := item.(type) {
    case *Context:
        p.Lock()
        defer p.Unlock()
        if _, ok := p.pvMap[*(item.mid)]; !ok {
            p.pvMap[*(item.mid)] = 0
        }

        p.pvMap[*(item.mid)] += 1

        log.Printf("parse item[%#v] Success\n", item)
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
    p.cleanUp()
}