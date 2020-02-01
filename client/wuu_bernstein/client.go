package wuu_bernstein

import (
	"Replicated-Blockchain/common"
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jpillora/backoff"
)

var (
	Client        *BlockClient
	clientMsgChan = make(chan (*common.ClientMessage))
	GlobalClock   = 0
	ClockLock     sync.Mutex
)

type BlockClient struct {
	ClientId int
	Peers    map[int]net.Conn
	TwoDTT   [][]int
	Log      *list.List
	Map      map[string]bool
}

// UpdateGlobalClock updates the global clock of the client.
// when local is set to true, it means that only the GlobalClock has to be updated. No comparision needed
func UpdateGlobalClock(ctx context.Context, currTimestamp int, clientId int, local bool) {
	ClockLock.Lock()
	defer ClockLock.Unlock()
	if local {
		GlobalClock += 1
		log.WithFields(log.Fields{
			"client_id": clientId,
			"clock":     GlobalClock,
		}).Info("updated the clock")
		return
	}
	if currTimestamp > GlobalClock {
		GlobalClock = currTimestamp + 1
	} else {
		GlobalClock += 1
	}
	log.WithFields(log.Fields{
		"client_id": clientId,
		"clock":     GlobalClock,
	}).Info("updated the clock")
}

func (client *BlockClient) processIncomingMessages(ctx context.Context) {
	for {
		select {
		case msg := <-clientMsgChan:
			log.WithFields(log.Fields{
				"log":            msg.Log,
				"Table":          msg.TwoDTT,
				"from_client_id": msg.FromId,
			}).Info("message received")
			UpdateGlobalClock(ctx, msg.Clock.Clock, client.ClientId, false)
			client.UpdateLog(ctx, msg.Log)
			client.UpdateFinalTable(ctx, msg.TwoDTT, client.ClientId-1, msg.FromId-1)
		default:
			continue
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
		log.Info("A new message received!")
		log.Info(resp.Log)
		log.Info(resp.TwoDTT)
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
	go client.listenToPeers(ctx)
	client.createConnectionTopology(ctx)
	go client.processIncomingMessages(ctx)
	go client.startUserInteractions(ctx)
	fmt.Println("done with process incoming messages")
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
		Map:      make(map[string]bool),
	}
}

func (client *BlockClient) PrintLog(ctx context.Context) string {
	var l string
	for block := client.Log.Front(); block != nil; block = block.Next() {
		l = l + strconv.Itoa(block.Value.(*common.Block).FromId) + ":" + strconv.Itoa(block.Value.(*common.Block).ToId) +
			":" + fmt.Sprintf("%g", block.Value.(*common.Block).Amount) +
			"[" + strconv.Itoa(block.Value.(*common.Block).Clock.Clock) + "]"
		if block.Next() != nil {
			l = l + "->"
		}
	}
	return l
}
