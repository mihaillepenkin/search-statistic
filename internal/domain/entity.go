package domain

type Message struct {
	Query        string `json:"query"`
	TimestampSec int64  `json:"timestamp_sec"`
}

type QueryCount struct {
	Query string `json:"query"`
	Count int    `json:"count"`
}