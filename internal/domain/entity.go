package domain

type Message struct {
	Query        string `json:"query"`
	TimestampSec int64  `json:"timestamp_sec"`
	UserID       string `json:"user_id,omitempty"`
	IP           string `json:"ip,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
}

type QueryCount struct {
	Query string `json:"query"`
	Count int    `json:"count"`
}