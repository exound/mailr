package main

import (
	"bufio"
	"encoding/json"
	"net"
)

type rejection interface {
	Rejection() string
}

type reject struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func (rej reject) Rejection() string {
	payload, _ := json.Marshal(rej)

	return string(payload)
}

type logons struct {
	Email  string `json:"email"`
}

type user struct {
	Nick   string `json:"nick"`
	Token  string `json:"token"`
	Logons logons `json:"logons"`
}

type request struct {
	Action string `json:"action"`
	User   user   `json:"user"`
}

const emailPattern = `^\w+[+-\.]?\w+@[a-z]+-?[a-z]+(\.[a-z]+-?[a-z]+)*\.[a-z]+$`

func newReject(reason, message string) reject {
	return reject{Reason: reason, Message: message}
}

func parseRequest(reqJSON string) (req request, rej rejection) {
	err := json.Unmarshal([]byte(reqJSON), &req)

	if err != nil {
		rej = newReject("user parsing error", err.Error())
	}

	return req, rej
}

func respond(conn net.Conn, res string) (int, error) {
	return conn.Write([]byte(res + "\n"))
}

func handleReq(conn net.Conn, s smtp) error {
	defer conn.Close()

	ok := `{"ok":true}`
	reader := bufio.NewReader(conn)

	message, err := reader.ReadString('\n')

	var rej rejection

	if err != nil {
		rej = newReject("reading error", "error reading request")
		_, err = respond(conn, rej.Rejection())
		return err
	}

	req, rej := parseRequest(message)

	if rej != nil {
		_, err = respond(conn, rej.Rejection())
		return err
	}

	rej = mail(req, s)

	if rej != nil {
		_, err = respond(conn, rej.Rejection())
		return err
	}

	_, err = respond(conn, ok)

	return err
}
