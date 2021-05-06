package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/adierkens/expo-server-sdk-go"
)

type ExpoPushMessage = expo.ExpoPushMessage

var expoClient = expo.NewExpo()
var token string

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func push_notif(token string, body string, title string, data map[string]interface{}) {
	m := expo.NewExpoPushMessage()

	m.To = token
	m.Body = body
	m.Title = title
	m.Data = data

	_, err := expoClient.SendPushNotification(m)
	check(err)
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	command := ""
	for _, s := range os.Args[1:] {
		command += s
		command += " "
	}

	// TODO : Setup token dir stuff if user doesn't have it

	homedir, err := os.UserHomeDir()
	check(err)
	token_file, err := os.Open(homedir + "/.clievent/expo_push_token")
	check(err)
	defer token_file.Close()

	reader := bufio.NewReader(token_file)
	token, err = reader.ReadString('\n')
	check(err)

	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	start := time.Now()
	err = cmd.Start()
	check(err)
	var ret_code string
	response := cmd.Wait()
	elapsed := time.Since(start)
	var body string
	if response == nil {
		ret_code = "0"
		body = "Successful!"
	} else {
		ret_code = response.Error()
		body = "Failed!"
	}
	hostname, err := os.Hostname()
	check(err)
	payload := make(map[string]interface{})
	payload["id"] = strconv.FormatInt(time.Now().Unix(), 10)
	payload["device"] = hostname
	payload["returned"] = ret_code
	payload["real_time"] = elapsed.String()
	push_notif(token[:41], body, command, payload)
}
