package common

import "net"

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
	FromId int     `json:"from_id"`
	ToId   int     `json:"to_id"`
	Amount float32 `json:"amount"`
}

type Peer struct {
	ClientId int
	Conn     net.Conn
}

type ClientResponse struct {
	Message string  `json:"message"`
	Balance float32 `json:"balance,omitempty"`
}

type ClientMessage struct {
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
	Amount     float32 `json:"amount,omitempty"`
	BalanceOf  int     `json:"balance_of,omitempty"`
}
