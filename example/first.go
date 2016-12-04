package main

import (
    "github.com/guyannanfei25/flowprocess"
    "fmt"
)

type Item struct {
    first1   string
    first2   string
    sencond  string
}

type FirstDispatch struct {
    flowprocess.DefaultDispatcher
}

func firstPreFunctor(item interface{}) {
    switch item := item.(type) {
    case *Item:
        item.first1 = "Hello"
        fmt.Println("FirstDispatch set item.first1 Success")
    default:
        fmt.Println("Not suitable item")
    }
}
