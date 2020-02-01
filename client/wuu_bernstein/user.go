package wuu_bernstein

import (
	"Replicated-Blockchain/common"
	"context"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/manifoldco/promptui"
)

// User interface
func (client *BlockClient) startUserInteractions(ctx context.Context) {
	var (
		err                       error
		receiverClient, amountStr string
		amount                    float64
		transactionType           string
		balanceOfClientId         int
		balanceOfClientStr        string
		receiverClientId          int
		message                   string
	)
	for {
		prompt := promptui.Select{
			Label: "Select Transaction",
			Items: []string{"Show Balance", "Transfer", "Send Message", "Print", "Exit"},
		}

		_, transactionType, err = prompt.Run()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error fetching transaction type from the command line")
			continue
		}
		log.WithFields(log.Fields{
			"choice": transactionType,
		}).Debug("You choose...")
		switch transactionType {
		case "Exit":
			log.Debug("Fun doing business with you, see you soon!")
			os.Exit(0)
		case "Show Balance":
			prompt := promptui.Prompt{
				Label: "Client Id",
			}
			balanceOfClientStr, err = prompt.Run()
			balanceOfClientId, _ = strconv.Atoi(balanceOfClientStr)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("error fetching the client number from the command line")
				continue
			}
			txn := &common.Txn{
				FromClient: client.ClientId,
				ToClient:   0,
				Type:       common.GetBalance,
				Amount:     0,
				BalanceOf:  balanceOfClientId,
				Message:    "",
				Clock: &common.LamportClock{
					PID:   client.ClientId,
					Clock: GlobalClock,
				},
			}
			client.ProcessEvent(ctx, txn)
		case "Transfer":
			prompt := promptui.Prompt{
				Label: "Receiver Client",
			}
			receiverClient, err = prompt.Run()
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("error fetching the client number from the command line")
				continue
			}
			receiverClientId, _ = strconv.Atoi(receiverClient)
			if receiverClientId == client.ClientId {
				log.Error("you cant send money to yourself!")
				continue
			}
			prompt = promptui.Prompt{
				Label:   "Amount to be transacted",
				Default: "",
			}
			amountStr, err = prompt.Run()
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("error fetching the transaction amount from the command line")
				continue
			}
			amount, _ = strconv.ParseFloat(amountStr, 64)
			txn := &common.Txn{
				FromClient: client.ClientId,
				ToClient:   receiverClientId,
				Type:       common.SendAmount,
				Amount:     amount,
				BalanceOf:  0,
				Message:    "",
				Clock: &common.LamportClock{
					PID:   client.ClientId,
					Clock: GlobalClock,
				},
			}
			client.ProcessEvent(ctx, txn)
		case "Send Message":
			prompt := promptui.Prompt{
				Label: "Receiver Client",
			}
			receiverClient, err = prompt.Run()
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("error fetching the client number from the command line")
				continue
			}
			receiverClientId, _ = strconv.Atoi(receiverClient)
			if receiverClientId == client.ClientId {
				log.Error("you cant send message to yourself!")
				continue
			}
			prompt = promptui.Prompt{
				Label: "Message",
			}
			message, err = prompt.Run()
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("error fetching the message from the command line")
				continue
			}
			txn := &common.Txn{
				FromClient: client.ClientId,
				ToClient:   receiverClientId,
				Type:       common.SendMessage,
				Amount:     0,
				BalanceOf:  0,
				Message:    message,
				Clock: &common.LamportClock{
					PID:   client.ClientId,
					Clock: GlobalClock,
				},
			}
			client.ProcessEvent(ctx, txn)
		case "Print":
			log.Info(common.Stars)
			log.WithFields(log.Fields{
				"log": client.PrintLog(ctx),
			}).Info("log")
			log.WithFields(log.Fields{
				"table": client.TwoDTT,
			}).Info("2dtt")
			log.Info(common.Stars)
		}
		//<-showNextPrompt
	}
}
