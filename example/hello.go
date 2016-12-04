package main

import (
    "github.com/guyannanfei25/flowprocess"
    "github.com/zieckey/goini"
    "fmt"
    "flag"
    "os"
    "time"
)

var framework flowprocess.DefaultDispatcher
var conf = flag.String("f", "conf.ini", "conf file path")

func main() {
    flag.Parse()
    ini, err := goini.LoadInheritedINI(*conf)
    if err != nil {
        fmt.Fprintf(os.Stderr, "parse conf[%s] err[%v]\n", *conf, err)
        os.Exit(-1)
    }

    framework.Init(ini, "")

    first1 := new(FirstDispatch)
    first2 := new(First2Dispatch)
    second := new(SecondDispatch)

    first1.Init(ini, "first")
    first2.Init(ini, "first2")
    second.Init(ini, "second")

    fmt.Printf("first1 name is %s\n", first1.GetName())
    fmt.Printf("first2 name is %s\n", first2.GetName())
    fmt.Printf("second name is %s\n", second.GetName())

    first1.SetPreFunc(firstPreFunctor)
    first2.SetPreFunc(first2PreFunctor)
    second.SetPreFunc(secondPreFunctor)
    second.SetSufFuc(secondSufFunctor)

    first1.DownRegister(second)
    first2.DownRegister(second)

    framework.DownRegister(first1)
    framework.DownRegister(first2)

    framework.Start()

    for i := 0; i < 1000; i++ {
        item := new(Item)
        framework.Dispatch(item)
    }

    time.Sleep(time.Second)

    framework.Close()
}
