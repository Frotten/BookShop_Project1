package models

const Capacity = 1000
const Workers = 3
const (
	RateOpNew int = 100 + iota
	RateOpUpdate
)
const TimeParseLayout = "2006-01-02 15:04:05"

var RateChan = make(chan *UserRateBook, Capacity) //通道容量先确认为1000
