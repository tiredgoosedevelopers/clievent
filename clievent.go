// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

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

func push_notif(token string, body string, title string) {
	// Make sure you set the `To`, `Title`, and `Body` properties
	m := expo.NewExpoPushMessage()

	m.To = token    // Your expo push token
	m.Body = body   // The body of the notification
	m.Title = title // The title of the notification

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
	if response == nil {
		ret_code = "0"
	} else {
		ret_code = response.Error()
	}
	hostname, err := os.Hostname()
	check(err)
	payload := strconv.Itoa(int(time.Now().Unix())) + ";" + hostname + ";" + ret_code + ";" + elapsed.String()
	push_notif(token[:41], payload, command)
	return
}
