package main

import (
	"fmt"
	"github.com/amzfans/appstarter/app"
	"github.com/amzfans/appstarter/domain"
	"log"
	"os"
	"os/signal"
)

const (
	stdout_socket_file = "stdout.sock"
	stderr_socket_file = "stderr.sock"
)

func main() {
	cmdAndArgs := os.Args[1:]
	if len(cmdAndArgs) == 0 {
		log.Fatal("Cannot get the runnable commands.")
	}

	cmd := cmdAndArgs[0]
	var args []string
	if len(cmdAndArgs) > 1 {
		args = cmdAndArgs[1:]
	}

	dsServer := domain.NewServer(stdout_socket_file, stderr_socket_file)
	err := dsServer.Start()
	if err != nil {
		panic(fmt.Sprintf("Cannot start the domain socket server as %s.", err.Error()))
	}
	defer dsServer.Stop()

	application := app.NewApp(dsServer, cmd, args...)
	err = application.Start()
	if err != nil {
		panic(fmt.Sprintf("Cannot start the application as %s.", err.Error()))
	}

	killChan := make(chan os.Signal)
	signal.Notify(killChan, os.Kill, os.Interrupt)
	for {
		select {
		case <-dsServer.NeedStop:
			shutdown(dsServer, application, nil)
		case <-application.NeedStop:
			shutdown(dsServer, application, nil)
		case osSig := <-killChan:
			shutdown(dsServer, application, osSig)
		}
	}

}

func shutdown(dsServer *domain.DomainSocketServer, application *app.Application, osSig os.Signal) {
	log.Println("Shutdown the domain socket server and kill the underline application.")
	dsServer.Stop()
	if !application.Cmd.ProcessState.Exited() {
		if osSig != nil {
			application.Cmd.Process.Signal(osSig)
		} else {
			application.Cmd.Process.Kill()
		}
	}
	os.Exit(0)
}
