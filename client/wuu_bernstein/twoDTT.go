package wuu_bernstein

import (
	"Replicated-Blockchain/common"
	"context"
	"math"
)

func (client *BlockClient) DecideLogForSending(ctx context.Context) []*common.Block {
	// iterate over all the logs in the current blockchain
	// collect all the logs which NEED to be transferred using the hasRecord Relationship
	// send the log and the timetable
	var (
		arr = make([]*common.Block, 0)
	)
	for block := client.Log.Front(); block != nil; block = block.Next() {
		if !client.HasRecord(ctx, block.Value.(*common.Block)) {
			arr = append(arr, block.Value.(*common.Block))
		}
	}
	return arr
}

func (client *BlockClient) UpdateTable(ctx context.Context) {
	ClockLock.Lock()
	defer ClockLock.Unlock()
	client.TwoDTT[client.ClientId-1][client.ClientId-1] = GlobalClock
}

func (client *BlockClient) HasRecord(ctx context.Context, block *common.Block) bool {
	// firstly, get the TT<client_id>[block.fromId, block.toId]
	// if this value is greater than the timestamp at which block event was registered, hasRecord is False
	if client.TwoDTT[block.FromId-1][block.ToId-1] > block.Clock.Clock {
		return false
	}
	return true
}

func (client *BlockClient) UpdateFinalTable(ctx context.Context, table [][]int, localRow int, remoteRow int) {
	for i, _ := range table {
		for j, _ := range table {
			client.TwoDTT[i][j] = int(math.Max(float64(client.TwoDTT[i][j]), float64(table[i][j])))
		}
	}
	// all the max values are updated, now update local row to the max of the 2 rows.
	for k, _ := range table {
		client.TwoDTT[localRow][k] = int(math.Max(float64(client.TwoDTT[localRow][k]), float64(table[remoteRow][k])))
	}
}
