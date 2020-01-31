package wuu_bernstein

import (
	"Replicated-Blockchain/common"
	"context"
	"fmt"

	log "github.com/Sirupsen/logrus"
)

func (client *BlockClient) SendAmount(ctx context.Context, request *common.Txn) (string, error) {
	var (
		block   *common.Block
		balance float64
	)
	if request.FromClient != client.ClientId {
		log.WithFields(log.Fields{
			"from_client_id":    request.FromClient,
			"to_client_id":      request.ToClient,
			"current_client_id": client.ClientId,
		}).Error("Sorry! You can only transfer money from your account!")
		return "", fmt.Errorf("Transfering money for another account initiated.")
	}
	balance, _ = client.GetBalance(ctx, &common.Txn{
		BalanceOf: request.BalanceOf,
		Type:      common.GetBalance,
	})
	if balance < request.Amount {
		return common.TxnIncorrect, nil
	}
	block = &common.Block{
		FromId: request.FromClient,
		ToId:   request.ToClient,
		Amount: request.Amount,
	}
	client.Log.PushBack(block)
	// TODO: update the 2D-TT??

	return common.TxnSuccess, nil
}

func (client *BlockClient) GetBalance(ctx context.Context, request *common.Txn) (float64, error) {
	var (
		balance float64
	)
	if client.Log.Front() == nil {
		return float64(10), nil
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
		balance float64
		err     error
		message string
	)
	// check the request, and based on it, take action
	common.UpdateGlobalClock(ctx, 0, client.ClientId, true)
	switch request.Type {
	case common.SendAmount:
		message, err = client.SendAmount(ctx, request)
		if err != nil {
			log.WithFields(log.Fields{
				"error":       err.Error(),
				"from_client": client.ClientId,
				"to_client":   request.ToClient,
			}).Error("error sending amount to client")
			return
		}
		log.Info("===========================================================")
		log.WithFields(log.Fields{
			"client_id": client.ClientId,
		}).Info(message)
		log.Info("===========================================================")
	case common.GetBalance:
		balance, _ = client.GetBalance(ctx, request)
		log.Info("===========================================================")
		log.WithFields(log.Fields{
			"client_id":         client.ClientId,
			"balance":           balance,
			"balance of client": request.BalanceOf,
		}).Info("BALANCE!")
		log.Info("===========================================================")
	case common.SendMessage:
		client.SendMessageToClients(ctx, request)
		log.Info("===========================================================")
		log.WithFields(log.Fields{
			"client_id":    client.ClientId,
			"to_client_id": request.ToClient,
		}).Info("MESSAGE SENT TO")
		log.Info("===========================================================")
	}
	<-showNextPrompt
}
