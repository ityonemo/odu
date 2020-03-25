package main

import (
	"os"
	"os/exec"
	"time"
)

func executor(workdir string, outputFifoPath string, args []string) error {
	proc := exec.Command(args[0], args[1:]...)
	proc.Dir = workdir

	logger.Printf("Command path: %v\n", proc.Path)

	outputFifo := openFifo(outputFifoPath, os.O_WRONLY)

	signal := make(chan bool)
	go startPipeline(proc, outputFifo, signal)

	// wait pipeline to start
	<-signal

	err := proc.Start()
	fatalIf(err)

	// wait for pipeline exit
	<-signal

	err = safeExit(proc)
	if e, ok := err.(*exec.Error); ok {
		// This shouldn't really happen in practice because we check for
		// program existence in Elixir, before launching odu
		logger.Printf("Run ERROR: %v\n", e)
		os.Exit(3)
	}
	// TODO: return Stderr and exit stauts to beam process
	logger.Printf("Run FINISHED: %#v\n", err)
	return err
}

func startPipeline(proc *exec.Cmd, outputFifo *os.File, signal chan bool) {
	// some commands expect stdin to be connected
	cmdInput, err := proc.StdinPipe()
	fatalIf(err)

	cmdOutput, err := proc.StdoutPipe()
	fatalIf(err)

	logger.Println("Starting pipeline")

	startInputConsumer(os.Stdin, cmdInput)
	outputStreamerExit := startOutputStreamer(cmdOutput, outputFifo)

	// signal that pipline is setup
	signal <- true

	// wait for pipline to exit
	<-outputStreamerExit

	// signal pipeline shutdown
	signal <- true
}

func safeExit(proc *exec.Cmd) error {
	done := make(chan error, 1)
	go func() {
		done <- proc.Wait()
	}()
	select {
	case <-time.After(3 * time.Second):
		if err := proc.Process.Kill(); err != nil {
			logger.Fatal("failed to kill process: ", err)
		}
		logger.Println("process killed as timeout reached")
		return nil
	case err := <-done:
		return err
	}
}

func openFifo(fifoPath string, mode int) *os.File {
	fifo, err := os.OpenFile(fifoPath, mode, 0600)
	if err != nil {
		fatal(err)
	}
	return fifo
}
