package app

import (
	"github.com/amzfans/appstarter/domain"
	"io"
	"log"
	"os/exec"
)

const buffer_size = 512

type Application struct {
	Cmd       *exec.Cmd
	cmdString string
	args      []string
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	dsServer  *domain.DomainSocketServer
}

func NewApp(dsServer *domain.DomainSocketServer, cmdString string, args ...string) *Application {
	return &Application{
		cmdString: cmdString,
		args:      args,
		dsServer:  dsServer,
	}
}

func (a *Application) Start() (err error) {
	a.Cmd = exec.Command(a.cmdString, a.args...)

	a.stdout, err = a.Cmd.StdoutPipe()
	if err != nil {
		return
	}

	a.stderr, err = a.Cmd.StderrPipe()
	if err != nil {
		return
	}

	err = a.Cmd.Start()
	if err != nil {
		return err
	}

	go a.writeStdDataToServer(false)
	go a.writeStdDataToServer(true)

	go func() {
		werr := a.Cmd.Wait()
		// for testing
		log.Println("The application command is completed.")
		if werr != nil {
			log.Printf("ERR: The cmd err is %s.", err.Error())
		}
	}()

	return
}

func (a *Application) writeStdDataToServer(isStderr bool) {
	rc := a.stdout
	dch := a.dsServer.StdoutDataChan
	if isStderr {
		rc = a.stderr
		dch = a.dsServer.StderrDataChan
	}

	defer rc.Close()
	buffer := make([]byte, buffer_size)
	for {
		count, err := rc.Read(buffer)
		if err != nil {
			log.Printf("ERR: Cannot read the data for the application's stdout or stderr as %s.", err.Error())
			return
		}

		select {
		case dch <- buffer[0:count]:
			// send the data to server's data channel.
		default:
			// just let it go.
			// for testing
			log.Printf("%v\n", buffer[0:count])
		}

	}
}
	
