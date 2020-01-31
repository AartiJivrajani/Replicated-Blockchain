package common

import (
	"net"
)

var (
	SendAmount  = "SEND AMOUNT"
	GetBalance  = "BALANCE"
	SendMessage = "SEND MESSAGE"

	// Transaction status
	TxnIncorrect = "INCORRECT"
	TxnSuccess   = "SUCCESS"
)
var ClientPortMap = map[int]int{
	1: 8000,
	2: 8001,
	3: 8002,
}

type Block struct {
	FromId  int     `json:"from_id"`
	ToId    int     `json:"to_id"`
	Amount  float64 `json:"amount"`
	Message string  `json:"message,omitempty"`
}

type Peer struct {
	ClientId int
	Conn     net.Conn
}

type ClientMessage struct {
	FromId  int           `json:"from_id"`
	ToId    int           `json:"to_id"`
	Log     []*Block      `json:"log"`
	Message string        `json:"message"`
	Clock   *LamportClock `json:"clock"`
}

type LogMessage struct {
}

type Log struct {
	LogList []*LogMessage
}

type TwoDTT struct {
}

type Txn struct {
	FromClient int     `json:"from_client"`
	ToClient   int     `json:"to_client"`
	Type       string  `json:"txn_type"`
	Amount     float64 `json:"amount,omitempty"`
	BalanceOf  int     `json:"balance_of,omitempty"`
	Message    string  `json:"message,omitempty"`
}

type LamportClock struct {
	PID   int `json:"pid"`
	Clock int `json:"clock"`
}
