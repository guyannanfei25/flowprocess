package main

import (
    "github.com/guyannanfei25/flowprocess"
    "fmt"
)

type SecondDispatch struct {
    flowprocess.DefaultDispatcher
}

func secondPreFunctor(item interface{}) {
    switch item := item.(type) {
    case *Item:
        item.sencond = "yc"
        fmt.Println("SecondDispatch set item.sencond Success")
    default:
        fmt.Println("Not suitable item")
    }

}

func secondSufFunctor(item interface{}) {
    fmt.Printf("Get item[%#v]\n", item)
}
