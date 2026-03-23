package models

const Capacity = 1000
const Workers = 3
const (
	RateOpNew int = 100 + iota
	RateOpUpdate
)

var RateChan = make(chan *UserRateBook, Capacity) //通道容量先确认为1000
