package wuu_bernstein

import (
	"Replicated-Blockchain/common"
	"context"
	"fmt"

	log "github.com/Sirupsen/logrus"
)

func (client *BlockClient) SendAmount(ctx context.Context, request *common.Txn) (*common.ClientResponse, error) {
	var (
		block   *common.Block
		balance float32
	)
	if request.FromClient != client.ClientId {
		log.WithFields(log.Fields{
			"from_client_id":    request.FromClient,
			"to_client_id":      request.ToClient,
			"current_client_id": client.ClientId,
		}).Error("Sorry! You can only transfer money from your account!")
		return nil, fmt.Errorf("Transfering money for another account initiated.")
	}
	balance, _ = client.GetBalance(ctx, &common.Txn{
		BalanceOf: request.BalanceOf,
		Type:      common.GetBalance,
	})
	if balance < request.Amount {
		return &common.ClientResponse{
			Message: common.TxnIncorrect,
		}, nil
	}
	block = &common.Block{
		FromId: request.FromClient,
		ToId:   request.ToClient,
		Amount: request.Amount,
	}
	client.Log.PushBack(block)
	return &common.ClientResponse{Message: common.TxnSuccess}, nil
}

func (client *BlockClient) GetBalance(ctx context.Context, request *common.Txn) (float32, error) {
	var (
		balance float32
	)
	if client.Log.Front() == nil {
		return float32(10), nil
	}
	balance = 10
	for block := client.Log.Front(); block != nil; block = block.Next() {
		if block.Value.(*common.Block).ToId == request.BalanceOf {
			balance += block.Value.(*common.Block).Amount
		}
		if block.Value.(*common.Block).FromId == request.BalanceOf {
			balance -= block.Value.(*common.Block).Amount
		}
	}
	return balance, nil
}

func (client *BlockClient) ProcessEvent(ctx context.Context, request *common.Txn) {
	var (
		balance        float32
		ClientResponse *common.ClientResponse
	)
	// check the request, and based on it, take action
	switch request.Type {
	case common.SendAmount:
		ClientResponse, _ = client.SendAmount(ctx, request)
	case common.GetBalance:
		balance, _ = client.GetBalance(ctx, request)
		ClientResponse = &common.ClientResponse{Message: "BALANCE", Balance: balance}
	case common.SendMessage:
		pass
	}
}
