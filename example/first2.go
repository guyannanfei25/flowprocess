package main

import (
    "github.com/guyannanfei25/flowprocess"
    "fmt"
)

type First2Dispatch struct {
    flowprocess.DefaultDispatcher
}

func first2PreFunctor(item interface{}) {
    switch item := item.(type) {
    case *Item:
        item.first2 = "World"
        fmt.Println("FirstDispatch set item.first2 Success")
    default:
        fmt.Println("Not suitable item")
    }
}
