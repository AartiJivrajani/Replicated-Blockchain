package wuu_bernstein

import (
	"Replicated-Blockchain/common"
	"container/list"
	"context"
	"encoding/json"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/manifoldco/promptui"

	log "github.com/Sirupsen/logrus"
	"github.com/jpillora/backoff"
)

var (
	Client         *BlockClient
	clientMsgChan  chan (*common.ClientMessage)
	GlobalClock    = 0
	ClockLock      sync.Mutex
	showNextPrompt = make(chan bool)
)

type BlockClient struct {
	ClientId int
	Peers    map[int]net.Conn
	TwoDTT   [][]int
	Log      *list.List
}

func (client *BlockClient) sendMessageOverWire(ctx context.Context, message *common.ClientMessage) {
	var (
		err      error
		jMessage []byte
	)
	jMessage, err = json.Marshal(message)
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err.Error(),
			"to_client": message.ToId,
			"message":   message,
		}).Error("error marshalling the general message")
		return
	}
	_, err = client.Peers[client.ClientId].Write(jMessage)
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err.Error(),
			"to_client": message.ToId,
		}).Error("error writing general message to the connection")
		return
	}
	return
}

func (client *BlockClient) SendMessageToClients(ctx context.Context, txn *common.Txn) {
	var (
		blockchain []*common.Block
	)
	blockchain = common.ListToArray(client.Log)
	clientMessage := &common.ClientMessage{
		FromId:  client.ClientId,
		ToId:    txn.ToClient,
		Log:     blockchain,
		Message: txn.Message,
		Clock: &common.LamportClock{
			PID:   client.ClientId,
			Clock: GlobalClock,
		},
	}
	client.sendMessageOverWire(ctx, clientMessage)
}

func (client *BlockClient) processIncomingMessages(ctx context.Context) {
	var (
		msg *common.ClientMessage
	)

	for {
		select {
		case msg = <-clientMsgChan:
			common.UpdateGlobalClock(ctx, msg.Clock.Clock, client.ClientId, false)
			// TODO:
			// 1. Update the current log
			// 2. Update the 2dtt
		}
	}
}
func (client *BlockClient) handleIncomingMessages(ctx context.Context, conn net.Conn) {
	var (
		err error
	)
	d := json.NewDecoder(conn)
	for {
		var resp *common.ClientMessage
		err = d.Decode(&resp)
		if err != nil {
			continue
		}
		// processIncomingMessages handles all the messages which this client receives on the wire
		clientMsgChan <- resp
	}
}

func (client *BlockClient) handleIncomingConnections(ctx context.Context, listener net.Listener) {
	var (
		conn net.Conn
		err  error
	)
	for {
		conn, err = listener.Accept()
		if err != nil {

		}
		go client.handleIncomingMessages(ctx, conn)
	}
}

func (client *BlockClient) listenToPeers(ctx context.Context) {
	var (
		listener net.Listener
		err      error
	)
	if listener, err = common.StartConnectionListener(client.ClientId); err != nil {
		return
	}
	go client.handleIncomingConnections(ctx, listener)
}

func (client *BlockClient) createConnectionTopology(ctx context.Context) {
	var (
		err  error
		conn net.Conn
		d    time.Duration
		b    = &backoff.Backoff{
			Min:    10 * time.Second,
			Max:    1 * time.Minute,
			Factor: 2,
			Jitter: true,
		}
	)
	for peer, _ := range client.Peers {
		PORT := ":" + strconv.Itoa(common.ClientPortMap[peer])
		d = b.Duration()
		for {
			conn, err = net.Dial("tcp", PORT)
			if err != nil {
				log.WithFields(log.Fields{
					"error":          err.Error(),
					"client_id":      client.ClientId,
					"peer_client_id": peer,
				}).Error("error connecting to the client")
				// if the connection fails, try to connect 3 times, post which just exit.
				if b.Attempt() <= 3 {
					time.Sleep(d)
					continue
				} else {
					log.Panic("unable to connect to the peers")
				}
			} else {
				log.WithFields(log.Fields{
					"client_id":      client.ClientId,
					"peer_client_id": peer,
				}).Debug("established connection with peer client")
				break
			}
		}
		client.Peers[peer] = conn
	}
}

func (client *BlockClient) Start(ctx context.Context) {
	// connect to the peers first before proceeding to the other tasks.
	client.createConnectionTopology(ctx)
	go client.listenToPeers(ctx)
	go client.startUserInteractions(ctx)
}

func NewClient(ctx context.Context, clientId int) *BlockClient {
	peers := make(map[int]net.Conn)
	if clientId == 1 {
		peers[2] = nil
		peers[3] = nil
	} else if clientId == 2 {
		peers[1] = nil
		peers[3] = nil
	} else if clientId == 3 {
		peers[1] = nil
		peers[2] = nil
	}
	twoTDTT := make([][]int, 3)
	for i := range twoTDTT {
		twoTDTT[i] = make([]int, 3)
	}
	// initialize the cells of 2-d timetable to 0
	for r, row := range twoTDTT {
		for c, _ := range row {
			twoTDTT[r][c] = 0
		}
	}
	return &BlockClient{
		ClientId: clientId,
		Peers:    peers,
		TwoDTT:   twoTDTT,
		Log:      list.New(),
	}
}

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
			Items: []string{"Show Balance", "Transfer", "Send Message", "Exit"},
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
			}
			client.ProcessEvent(ctx, txn)
		}
		<-showNextPrompt
	}
}
