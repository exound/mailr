package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func listen(path string) {
	s, err := parseSMTP()

	println(fmt.Sprintf("Listening socket on path: %s", path))

	if err != nil {
		panic(err)
	}

	ln, _ := net.Listen("unix", path)

	defer ln.Close()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)

	go func(c chan os.Signal) {
		sig := <-c

		fmt.Println(fmt.Sprintf("Caught signal %s: shutting down.", sig))

		ln.Close()

		os.Exit(0)
	}(sigc)

	conns := make(chan net.Conn)
	errs := make(chan error)

	go func() {
		for {
			conn, _ := ln.Accept()

			conns <- conn
		}
	}()

	for {
		select {
		case conn := <-conns:
			go func(conn net.Conn) {
				defer conn.Close()

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
