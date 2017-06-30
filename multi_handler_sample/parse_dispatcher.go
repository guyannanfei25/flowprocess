package main

import (
    "log"
    "strings"
    "strconv"
    sj  "github.com/bitly/go-simplejson"
)

type ParseDispatcher struct {
}

func (p *ParseDispatcher) Init(conf *sj.Json) error {
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

func (p *ParseDispatcher) Close() {
}
