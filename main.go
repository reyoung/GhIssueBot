package main

import (
	"fmt"
	"github.com/bmatsuo/go-jsontree"
	"github.com/jessevdk/go-flags"
	"github.com/reyoung/hookserve/hookserve"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"time"
)

func e(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type HttpOptions struct {
	Port int
}

type DutyTable struct {
	Sun   []string
	Mon   []string
	Tue   []string
	Wed   []string
	Thurs []string
	Fri   []string
	Sat   []string
}

type EmailAccount struct {
	Addr     string
	Password string
}

type Options struct {
	Http       HttpOptions
	SecretCode *string
	Duty       DutyTable
	Email      EmailAccount
}

func newOptions() *Options {
	return &Options{
		Http: HttpOptions{
			Port: 8000,
		},
		SecretCode: nil,
		Duty: DutyTable{
			Sun:   []string{},
			Mon:   []string{},
			Tue:   []string{},
			Wed:   []string{},
			Thurs: []string{},
			Fri:   []string{},
			Sat:   []string{},
		},
		Email: EmailAccount{},
	}
}

func parseOpts() *Options {
	var opts struct {
		Config string `short:"c" long:"config" description:"configuration file" required:"true"`
	}
	_, err := flags.Parse(&opts)

	printDefaultOptions := func() {
		opt := newOptions()
		buf, err := yaml.Marshal(opt)
		e(err)
		fmt.Println(string(buf[:]))
	}

	if err != nil {
		fmt.Println("Parsing command line argument error.")
		fmt.Println("The default configuration file is ")
		printDefaultOptions()
		e(err)
	}

	f, err := os.Open(opts.Config)
	if err != nil {
		fmt.Printf("Cannot open config file %s", opts.Config)
		e(err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	e(err)
	o := newOptions()
	e(yaml.Unmarshal(buf, o))
	return o
}

func main() {
	opts := parseOpts()
	server := hookserve.NewServer()
	server.CustomEventHandler["issues"] = on_issue_hook
	server.CustomEventHandler["issue_comment"] = on_issue_comment_hook
	server.Port = opts.Http.Port
	if opts.SecretCode != nil {
		server.Secret = *opts.SecretCode
	}
	server.GoListenAndServe()
	dutyTable := [][]string{
		opts.Duty.Sun,
		opts.Duty.Mon,
		opts.Duty.Tue,
		opts.Duty.Wed,
		opts.Duty.Thurs,
		opts.Duty.Fri,
		opts.Duty.Sat,
	}

	for_each_duty := func(callback func(string)) {
		dutyMember := dutyTable[time.Now().Weekday()]
		for _, each_member := range dutyMember {
			callback(each_member)
		}
	}

	for event := range server.Events {
		switch event.(type) {
		case hookserve.Event:
			log.Println("Hookserve Event Recieved")
		case *IssueEvent:
			{
				issue := event.(*IssueEvent)
				if issue.action == "opened" || issue.action == "reopen" || issue.action == "created" {
					for_each_duty(func(to string) {
						send(fmt.Sprintf("[GITHUB ISSUE] %s %s", issue.title, issue.action),
							fmt.Sprintf(`Today is on your duty to handle github issues,
%s is %s.
URL: %s`, issue.title, issue.action, issue.url), opts.Email.Addr, to, opts.Email.Password)
					})
				}
			}
		case *IssueCommentEvent:
			{
				comment := event.(*IssueCommentEvent)
				if comment.action == "created" {
					for_each_duty(func(to string) {
						send(fmt.Sprintf("[GITHUB ISSUE_COMMENTS] %s %s comments on %s", comment.user, comment.action, comment.issue_title),
							fmt.Sprintf(`Today is on your duty to handle github issues,
%s %s comment on %s,
URL: %s
---
%s`, comment.user, comment.action, comment.issue_title, comment.url, comment.body), opts.Email.Addr, to, opts.Email.Password)
					})
				}
			}

		}
	}
}

type IssueEvent struct {
	action string
	url    string
	title  string
}

func on_issue_hook(request *jsontree.JsonTree) (ev interface{}, err error) {
	event := &IssueEvent{}
	ev = event
	event.action, err = request.Get("action").String()
	if err != nil {
		return
	}
	event.url, err = request.Get("issue").Get("html_url").String()
	if err != nil {
		return
	}
	event.title, err = request.Get("issue").Get("title").String()
	return
}

type IssueCommentEvent struct {
	action      string
	user        string
	issue_title string
	body        string
	url	    string
}

func on_issue_comment_hook(request *jsontree.JsonTree) (ev interface{}, err error) {
	event := &IssueCommentEvent{}
	ev = event
	event.action, err = request.Get("action").String()
	if err != nil {
		return
	}
	event.user, err = request.Get("comment").Get("user").Get("login").String()
	if err != nil {
		return
	}
	event.issue_title, err = request.Get("issue").Get("title").String()
	if err != nil {
		return
	}
	event.body, err = request.Get("comment").Get("body").String()
	if err != nil {
		return
	}
	event.url, err = request.Get("comment").Get("html_url").String()
	return
}

func send(title string, body string, from string, to string, pass string) {
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		fmt.Sprintf("Subject: %s\n\n", title) +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))
	e(err)
}
