package wuu_bernstein

import (
	"Replicated-Blockchain/common"
	"container/list"
	"context"
	"encoding/json"
	"net"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jpillora/backoff"
)

var (
	Client        *BlockClient
	clientMsgChan chan (*common.ClientMessage)
)

type BlockClient struct {
	ClientId int
	Q        *list.List
	Peers    map[int]net.Conn
	TwoDTT   [][]int
	Log      *list.List
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
		Q:        list.New(),
		Peers:    peers,
		TwoDTT:   twoTDTT,
		Log:      list.New(),
	}
}
