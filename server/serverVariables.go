package server

import (
	"time"
)

var OrderTickers = make(map[int]*time.Ticker)
