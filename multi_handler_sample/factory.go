package main

import (
    "github.com/guyannanfei25/flowprocess"
)

func CreatorFactory(ctype string) flowprocess.HandlerCreator {
    switch ctype {
    case "filter":
        return func() flowprocess.Handler {
            return new(FilterDispatcher)
        }
    case "parse":
        return func() flowprocess.Handler {
            return new(ParseDispatcher)
        }
    case "pv":
        return func() flowprocess.Handler {
            return new(PvDispatcher)
        }
    default:
        return nil
    }
}
