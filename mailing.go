package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/gomail.v2"
	"html/template"
	"regexp"
)

type smtp struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Port     int    `json:"port"`
}

type email struct {
	To, Subject, Body string
}

func parseSMTP() (s smtp, err error) {
	data, err := Asset("smtp.json")

	if err != nil {
		return s, err
	}

	err = json.Unmarshal(data, &s)

	if err != nil {
		return s, errors.New("smtp parsing error: " + err.Error())
	}

	return s, err
}

func emailFromReq(req request) (em email, rej rejection) {
	actionToSubject := map[string]string{
		"reset-password": "叉烧网 - 重置密码",
	}

	action := req.Action
	subject := actionToSubject[action]
	if subject == "" {
		return em, newReject("invalid data", "invalid action")
	}

	u := req.User
	if u.Nick == "" || u.Logons.Email == "" {
		return em, newReject("invalid data", "nick and email should not be empty")
	}

	if action == "reset-password" && u.Token == "" {
		return em, newReject("invalid data", "token must be present to reset-password")
	}

	validEmail := regexp.MustCompile(emailPattern)

	if !validEmail.MatchString(u.Logons.Email) {
		return em, newReject("invalid data", "invalid email address")
	}

	byt, _ := Asset(fmt.Sprintf("templates/%s.html", action))

	tpl, _ := template.New(action).Parse(string(byt))

	var buf bytes.Buffer

	tpl.Execute(&buf, u)

	em = email{To: u.Logons.Email, Subject: subject, Body: buf.String()}

	return em, rej
}

func mail(req request, s smtp) (rej rejection) {
	em, rej := emailFromReq(req)

	if rej != nil {
		return rej
	}

	s, err := parseSMTP()

	if err != nil {
		panic(err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.User)
	m.SetHeader("To", em.To)
	m.SetHeader("Subject", em.Subject)
	m.SetBody("text/html", em.Body)

	d := gomail.NewDialer(s.Host, s.Port, s.User, s.Password)

	err = d.DialAndSend(m)

	if err != nil {
		return newReject("sending error", err.Error())
	}

	return rej
}
