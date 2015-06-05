package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	. "github.com/tnantoka/chatsworth"
	"io/ioutil"
	"log"
	"math"
	"strings"
	"time"
)

func main() {
	var f = flag.String("f", "./hosts", "")
	var k = flag.String("k", "./.api_token", "")
	var r = flag.String("r", "", "")
	flag.Parse()

	cw := Chatsworth{
		RoomID:   *r,
		APIToken: loadToken(*k),
	}
	cw.PostMessage(buildMessage(*f))
}

func loadToken(file string) string {
	token, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	return string(token)
}

const layout = "2006年1月2日15時04分"

func buildMessage(file string) string {
	hosts, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	var validHosts []string
	for _, host := range strings.Split(string(hosts), "\n") {
		if len(host) > 0 {
			validHosts = append(validHosts, host)
		}
	}

	messageChan := fetchTimes(validHosts)

	message := "[info][title]SSL証明書の期限[/title]"
	for i := 0; i < len(validHosts); i++ {
		m := <-messageChan
		fmt.Print(m)
		message += m
	}
	message += "[/info]"

	return message
}

const warningDays = 30

func fetchTimes(hosts []string) <-chan string {
	messageChan := make(chan string)

	jst, _ := time.LoadLocation("Asia/Tokyo")
	for _, host := range hosts {
		go func(host string) {
			t := fetchTime(host).In(jst)
			duration := t.Sub(time.Now())
			days := math.Floor(duration.Hours() / 24)
			message := host + ": " + t.Format(layout)
			message += "（あと" + fmt.Sprint(days) + "日）"
			if days < warningDays {
				message += " (devil) "
			}
			message += "\n"
			messageChan <- message
		}(host)
	}

	return messageChan
}

func fetchTime(host string) time.Time {
	config := tls.Config{}

	conn, err := tls.Dial("tcp", host+":443", &config)
	if err != nil {
		log.Fatal("host: " + host + ", error: " + err.Error())
	}

	state := conn.ConnectionState()
	certs := state.PeerCertificates

	defer conn.Close()

	return certs[0].NotAfter
}
