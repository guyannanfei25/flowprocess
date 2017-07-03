package main

import (
    sj  "github.com/guyannanfei25/go-simplejson"
    "log"
    "sync"
)

type PvDispatcher struct {
    pvMap    map[string]int
    id       int
    sync.Mutex
}

func (p *PvDispatcher) Init(conf *sj.Json, id int) error {
    p.pvMap = make(map[string]int)
    p.id = id
    log.Printf("%dth PvDispatcher init\n", id)

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

func (p *PvDispatcher) Tick() {
}

func (p *PvDispatcher) cleanUp() {
    for mid, pv := range p.pvMap {
        log.Printf("%s pv is %d\n", mid, pv)
    }
}

func (p *PvDispatcher) Close() {
    p.cleanUp()
    log.Printf("%dth PvDispatcher close\n", p.id)
}
