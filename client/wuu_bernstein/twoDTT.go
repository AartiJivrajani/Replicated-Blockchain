package wuu_bernstein

import (
	"Replicated-Blockchain/common"
	"context"
	"fmt"
	"math"

	log "github.com/Sirupsen/logrus"
)

func (client *BlockClient) UpdateLog(ctx context.Context, newLog []*common.Block) {
	log.Info(common.Stars)
	log.WithFields(log.Fields{
		"log": client.PrintLog(ctx),
	}).Info("Log - before update")
	for _, block := range newLog {
		if _, OK := client.Map[fmt.Sprintf("%s-%s", block.Clock.PID, block.Clock.Clock)]; !OK {
			client.Log.PushBack(block)
			client.Map[fmt.Sprintf("%s-%s", block.Clock.PID, block.Clock.Clock)] = true
		}
	}
	log.WithFields(log.Fields{
		"log": client.PrintLog(ctx),
	}).Info("Log - after update")
	log.Info(common.Stars)
}

func (client *BlockClient) DecideLogForSending(ctx context.Context, receiverId int) []*common.Block {
	// iterate over all the logs in the current blockchain
	// collect all the logs which NEED to be transferred using the hasRecord Relationship
	// send the log and the timetable
	var (
		arr = make([]*common.Block, 0)
	)
	log.WithFields(log.Fields{
		"log":   client.PrintLog(ctx),
		"table": client.TwoDTT,
	}).Info("log/table before deciding")
	for block := client.Log.Front(); block != nil; block = block.Next() {
		if !client.HasRecord(ctx, block.Value.(*common.Block), receiverId) {
			arr = append(arr, block.Value.(*common.Block))
		}
	}
	log.WithFields(log.Fields{
		"log": arr,
	}).Debug("log after deciding")
	return arr
}

func (client *BlockClient) UpdateTable(ctx context.Context) {
	ClockLock.Lock()
	defer ClockLock.Unlock()
	client.TwoDTT[client.ClientId-1][client.ClientId-1] = GlobalClock
}

func (client *BlockClient) HasRecord(ctx context.Context, block *common.Block, receiverId int) bool {
	// if this value is greater than the timestamp at which block event was registered, hasRecord is False
	log.WithFields(log.Fields{
		"row": receiverId - 1,
		"col": block.EventSourceId - 1,
	}).Info("fetching info for: ")
	if client.TwoDTT[receiverId-1][block.EventSourceId-1] >= block.Clock.Clock {
		log.WithFields(log.Fields{
			"value": client.TwoDTT[receiverId-1][block.EventSourceId-1],
			"clock": block.Clock.Clock,
		}).Info("returning true")
		return true
	}
	log.WithFields(log.Fields{
		"value": client.TwoDTT[receiverId-1][block.EventSourceId-1],
		"clock": block.Clock.Clock,
	}).Info("returning false")
	return false
}

func (client *BlockClient) UpdateFinalTable(ctx context.Context, table [][]int, localRow int, remoteRow int) {
	log.Info(common.Stars)
	log.WithFields(log.Fields{
		"table": client.TwoDTT,
	}).Info("2d-TT - Before update")
	for i, _ := range table {
		for j, _ := range table {
			client.TwoDTT[i][j] = int(math.Max(float64(client.TwoDTT[i][j]), float64(table[i][j])))
		}
	}
	// all the max values are updated, now update local row to the max of the 2 rows.
	for k, _ := range table {
		client.TwoDTT[localRow][k] = int(math.Max(float64(client.TwoDTT[localRow][k]), float64(table[remoteRow][k])))
	}
	log.WithFields(log.Fields{
		"table": client.TwoDTT,
	}).Info("2d-TT - After update")
	log.Info(common.Stars)
}
