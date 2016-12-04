package main

import (
    "github.com/guyannanfei25/flowprocess"
    "github.com/zieckey/goini"
    "fmt"
    "flag"
    "os"
    "strings"
    // "time"
    "bufio"
)

var framework flowprocess.DefaultDispatcher
var conf = flag.String("f", "conf.ini", "conf file path")
var logPath  = flag.String("l", "test.log", "log file path")

func main() {
    flag.Parse()
    ini, err := goini.LoadInheritedINI(*conf)

    if err != nil {
        fmt.Fprintf(os.Stderr, "parse conf[%s] err[%v]\n", *conf, err)
        os.Exit(-1)
    }

    framework.Init(ini, "")

    parse  := new(ParseDispatcher)
    pv     := new(PvDispatcher)
    filter := new(FilterDispatcher)

    parse.Init(ini, "parse_dispatcher")
    // pv.Initialize(ini, "pv_dispatcher")
    pv.Init(ini, "pv_dispatcher")
    filter.Initialize(ini, "filter_dispatcher")

    parse.SetPreFunc(parseline)
    pv.SetPreFunc(pv.pvParse)
    filter.SetPreFunc(filter.filter)

    pv.SetCleanUp(pv.cleanUp)
    filter.SetCleanUp(filter.cleanUp)

    parse.DownRegister(pv)
    parse.DownRegister(filter)

    framework.DownRegister(parse)

    framework.Start()


    // 业务逻辑

    fp, err := os.Open(*logPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Open log[%s] err[%v]\n", *logPath, err)
        os.Exit(-2)
    }
    defer fp.Close()

    scanner := bufio.NewScanner(fp)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.Index(line, "#") != -1 {
            continue
        }
        fmt.Printf("Process line[%s]\n", line)
        item := new(Context)
        item.raw = &line
        framework.Dispatch(item)
    }

    framework.Close()
}
