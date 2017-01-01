package main

import (
    "github.com/guyannanfei25/flowprocess"
    "log"
    "strings"
    "strconv"
)

type ParseDispatcher struct {
    flowprocess.DefaultDispatcher
}

func parseline(item interface{}) (interface{}, error) {
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
