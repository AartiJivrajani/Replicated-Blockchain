package common

import (
	"fmt"
	"net"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

func StartConnectionListener(clientId int) (net.Listener, error) {
	var (
		PORT     string
		err      error
		listener net.Listener
	)
	PORT = ":" + strconv.Itoa(ClientPortMap[clientId])
	listener, err = net.Listen("tcp", PORT)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"client_id":   clientId,
			"client_port": ClientPortMap[clientId],
		}).Error("error starting a listener on the port")
		return nil, fmt.Errorf("error starting a listener on the port")
	}
	return listener, nil
}
