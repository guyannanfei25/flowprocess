package main

import (
    "log"
    "strings"
    "strconv"
    sj  "github.com/guyannanfei25/go-simplejson"
)

type ParseDispatcher struct {
    id               int
}

func (p *ParseDispatcher) Init(conf *sj.Json, id int) error {
    p.id   = id
    log.Printf("%dth ParseDispatcher init\n", id)

    return nil
}

func (p *ParseDispatcher) Process(item interface{}) (interface{}, error) {
    switch item := item.(type) {
    case *Context:
        fields := strings.Split(*item.raw, "\t")
        item.mid = &fields[0]
        item.level, _ = strconv.Atoi(fields[1])

        log.Printf("parse line[%s] to item[%#v] Success\n", *item.raw, item)
    default:
        log.Printf("Get unknown type[%T]\n", item)
    }
    return item, nil
}

func (p *ParseDispatcher) Tick() {
}

func (p *ParseDispatcher) Close() error {
    log.Printf("%dth ParseDispatcher close\n", p.id)
    return nil
}
