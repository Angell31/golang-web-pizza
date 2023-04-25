package server

import (
	"pizza/data"
	"time"
)

var MySessionName *data.User
var OrderTickers = make(map[int]*time.Ticker)
