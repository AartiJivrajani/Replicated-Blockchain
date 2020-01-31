package common

import (
	"container/list"
)

var (
	Stars = "*********************************************************************************************"
	Dashes = "============================================================================================="
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
			Clock:   block.Value.(*Block).Clock,
		}
		arr = append(arr, arrBlock)
	}
	return arr
}
