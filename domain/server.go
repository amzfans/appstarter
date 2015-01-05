package domain

import (
	"errors"
	"log"
	"net"
)

type DomainSocketServer struct {
	stdoutPath     string
	stderrPath     string
	StdoutDataChan chan []byte
	StderrDataChan chan []byte
	NeedStop       chan bool
	stopping       chan bool
}

func NewServer(stdoutPath, stderrPath string) *DomainSocketServer {
	return &DomainSocketServer{
		stdoutPath:     stdoutPath,
		stderrPath:     stderrPath,
		StdoutDataChan: make(chan []byte),
		StderrDataChan: make(chan []byte),
		NeedStop:       make(chan bool),
		stopping:       make(chan bool),
	}
}

func (dsServer *DomainSocketServer) Start() (err error) {
	if dsServer.stdoutPath == "" || dsServer.stderrPath == "" {
		err = errors.New("The stdout and stder socket file paths must be assigned.")
		return
	}

	stdoutListener, err := net.Listen("unix", dsServer.stdoutPath)
	if err != nil {
		return
	}
	stderrListener, err := net.Listen("unix", dsServer.stderrPath)
	if err != nil {
		return
	}

	go dsServer.forwardToClient(stdoutListener, false)
	go dsServer.forwardToClient(stderrListener, true)

	return
}

func (dsServer *DomainSocketServer) forwardToClient(ls net.Listener, isStdErr bool) {
	defer ls.Close()
	// only accept one client.
	conn, err := ls.Accept()
	if err != nil {
		log.Printf("ERR: Cannot accept the client connection as %s.", err.Error())
		dsServer.NeedStop <- true
		return
	}

	var dataChan chan []byte
	if isStdErr {
		dataChan = dsServer.StderrDataChan
	} else {
		dataChan = dsServer.StdoutDataChan
	}

	for byteData := range dataChan {
		select {
		case <-dsServer.stopping:
			return
		default:
			_, werr := conn.Write(byteData)
			if werr != nil {
				log.Printf("ERR: Cannot send the data via client connection as %s.", werr.Error())
				dsServer.NeedStop <- true
				return
			}
		}

	}
}

func (dsServer *DomainSocketServer) Stop() {
	dsServer.stopping <- true
	close(dsServer.StdoutDataChan)
	close(dsServer.StderrDataChan)
}
