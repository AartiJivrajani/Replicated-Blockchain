package common

import (
	"Replicated-Blockchain/client/wuu_bernstein"
	"container/list"
	"context"

	log "github.com/Sirupsen/logrus"
)

func ListToArray(l *list.List) []*Block {
	var (
		arr = make([]*Block, 0)
	)
	for block := l.Front(); block != nil; block = block.Next() {
		arrBlock := &Block{
			FromId:  block.Value.(*Block).FromId,
			ToId:    block.Value.(*Block).ToId,
			Amount:  block.Value.(*Block).Amount,
			Message: block.Value.(*Block).Message,
		}
		arr = append(arr, arrBlock)
	}
	return arr
}

// UpdateGlobalClock updates the global clock of the client.
// when local is set to true, it means that only the GlobalClock has to be updated. No comparision needed
func UpdateGlobalClock(ctx context.Context, currTimestamp int, clientId int, local bool) {
	wuu_bernstein.ClockLock.Lock()
	defer wuu_bernstein.ClockLock.Unlock()
	if local {
		wuu_bernstein.GlobalClock += 1
		return
	}
	if currTimestamp > wuu_bernstein.GlobalClock {
		wuu_bernstein.GlobalClock = currTimestamp + 1
	} else {
		wuu_bernstein.GlobalClock += 1
	}
	log.WithFields(log.Fields{
		"client_id": clientId,
		"clock":     wuu_bernstein.GlobalClock,
	}).Info("updated the clock")
}
