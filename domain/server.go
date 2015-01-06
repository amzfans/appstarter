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
	stdoutListener net.Listener
	stderrListener net.Listener
}

func NewServer(stdoutPath, stderrPath string) *DomainSocketServer {
	return &DomainSocketServer{
		stdoutPath:     stdoutPath,
		stderrPath:     stderrPath,
		StdoutDataChan: make(chan []byte),
		StderrDataChan: make(chan []byte),
		NeedStop:       make(chan bool, 1),
		stopping:       make(chan bool, 1),
	}
}

func (dsServer *DomainSocketServer) Start() (err error) {
	if dsServer.stdoutPath == "" || dsServer.stderrPath == "" {
		err = errors.New("The stdout and stder socket file paths must be assigned.")
		return
	}

	dsServer.stdoutListener, err = net.Listen("unix", dsServer.stdoutPath)
	if err != nil {
		return
	}
	dsServer.stderrListener, err = net.Listen("unix", dsServer.stderrPath)
	if err != nil {
		return
	}

	go dsServer.forwardToClient(false)
	go dsServer.forwardToClient(true)

	return
}

func (dsServer *DomainSocketServer) forwardToClient(isStdErr bool) {
	ls := dsServer.stdoutListener
	dataChan := dsServer.StdoutDataChan
	if isStdErr {
		ls = dsServer.stderrListener
		dataChan = dsServer.StderrDataChan
	}
	defer ls.Close()
	// only accept one client.
	log.Println("Waiting for client...")
	conn, err := ls.Accept()
	log.Println("Accept the client or get the erro.")
	if err != nil {
		log.Printf("ERR: Cannot accept the client connection as %s.", err.Error())
		select {
		case dsServer.NeedStop <- true:
			return
		default:
			log.Println("Return for erro.")
			return	
		}
	}
	defer conn.Close()
	for {
		select {
	        case byteData := <-dataChan:
        	        _, werr := conn.Write(byteData)
	                if werr != nil {
                        	log.Printf("ERR: Cannot send the data via client connection as %s.", werr.Error())
                        	select {
				case dsServer.NeedStop <- true:
					return
				default:
					return
				}
        	        }
		case <-dsServer.stopping:
			log.Println("The socket server is stopping.")
			return
	        }

	}
}

func (dsServer *DomainSocketServer) Stop() {
	dsServer.stopping <- true
	if dsServer.stdoutListener != nil {
		dsServer.stdoutListener.Close()
	}
	if dsServer.stderrListener != nil {
		dsServer.stderrListener.Close()
	}
	close(dsServer.StdoutDataChan)
	close(dsServer.StderrDataChan)
}
