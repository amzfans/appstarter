package main

import (
	"fmt"
	"github.com/johnxiaoyi/appstarter/app"
	"github.com/johnxiaoyi/appstarter/domain"
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
			log.Println("Shutdown the domain socket server and kill the underline application.")
			dsServer.Stop()
			application.Cmd.Process.Kill()
		case osSig := <-killChan:
			log.Println("Shutdown the domain socket server and the underline application.")
			dsServer.Stop()
			application.Cmd.Process.Signal(osSig)
			return
		}
	}

}
