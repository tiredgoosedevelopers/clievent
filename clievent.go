// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"bufio"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/adierkens/expo-server-sdk-go"
	"github.com/afdalwahyu/gonnel"
	"github.com/gorilla/websocket"
)

type ExpoPushMessage = expo.ExpoPushMessage

var expoClient = expo.NewExpo()
var token string
var addr = flag.String("addr", "localhost:8080", "http service address")
var upgrader = websocket.Upgrader{} // use default options

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func push_notif(token string, body string, title string) {
	// Make sure you set the `To`, `Title`, and `Body` properties
	m := expo.NewExpoPushMessage()

	m.To = token    // Your expo push token
	m.Body = body   // The body of the notification
	m.Title = title // The title of the notification

	_, err := expoClient.SendPushNotification(m)
	check(err)

	/*
		// Full API
		expo.NewExpo()            // create a new client
		expo.NewExpoPushMessage() // create a msg to push
		var messages []*ExpoPushMessage
		expo.ChunkPushNotifications(messages) // Chunk them into an array of batches

		client.SendPushNotification(messages[0]) // Send a single message
		client.SendPushNotifications(messages)   // Send a batch of messages
	*/
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

	ngrok, err := gonnel.NewClient(gonnel.Options{
		BinaryPath: "/Users/mohammadyazdani/bin/ngrok", // Change to dynamic path
	})
	if err != nil {
		fmt.Println(err)
	}
	defer ngrok.Close()
	done := make(chan bool)
	go ngrok.StartServer(done)
	<-done

	ngrok.AddTunnel(&gonnel.Tunnel{
		Proto:        gonnel.TCP,
		Name:         "clievent",
		LocalAddress: "127.0.0.1:8080",
	})

	ngrok.ConnectAll()

	// TODO : Uncomment for websocket stuff
	// push_notif(token[:41], ngrok.Tunnel[0].RemoteAddress[6:], "Tunnel address")

	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	check(err)
	ret_code := cmd.Wait()
	if ret_code == nil {
		push_notif(token[:41], "Successful!", command)
	} else {
		push_notif(token[:41], "Failed!", command)
	}
	ngrok.DisconnectAll()
	return

	/* TODO : Websocket stuff
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
	ngrok.DisconnectAll()
	*/
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
        output.scroll(0, output.scrollHeight);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output" style="max-height: 70vh;overflow-y: scroll;"></div>
</td></tr></table>
</body>
</html>
`))
