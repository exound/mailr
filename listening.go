package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type closable interface {
	Close() error
}

func closeClosable(ln closable) {
	err := ln.Close()

	if err != nil {
		log.Fatal(err)
	}
}

func cleanupSocket(socketPath string) {
	if _, err := os.Stat(socketPath); err == nil {
		if err = os.Remove(socketPath); err != nil {
			log.Fatal(err)
		}
	}
}

func tearDown(c chan os.Signal, socketPath string, ln closable) {
	sig := <-c

	fmt.Println(fmt.Sprintf("\nCaught signal %s: shutting down.", sig))
	closeClosable(ln)
	cleanupSocket(socketPath)

	os.Exit(0)
}

func listen(path string) {
	cleanupSocket(path)

	s, err := parseSMTP()

	println(fmt.Sprintf("Listening socket on path: %s", path))

	if err != nil {
		panic(err)
	}

	ln, err := net.Listen("unix", path)

	defer closeClosable(ln)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, os.Kill, syscall.SIGTERM)

	go tearDown(signalChannel, path, ln)

	connectionChannel := make(chan net.Conn)
	errs := make(chan error)

	go func() {
		for {
			conn, _ := ln.Accept()

			connectionChannel <- conn
		}
	}()

	for {
		select {
		case conn := <-connectionChannel:
			go func(conn net.Conn) {
				defer closeClosable(conn)

				err := handleReq(conn, s)

				if err != nil {
					errs <- err
				}
			}(conn)
		case err := <-errs:
			println(err.Error())

			break
		}
	}
}
