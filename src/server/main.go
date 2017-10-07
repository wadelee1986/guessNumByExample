// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	blockchain "github.com/wadelee1986/guessNumByExample/src/blockchainSDK"

	"github.com/sirupsen/logrus"
)

var addr = flag.String("addr", ":8080", "http service address")

func StaticServer(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "GET,POST, DELETE")
	fmt.Println("Static server begin ...........")
	dir := "/home/wade.lee/goWorkProject/src/github.com/wadelee1986/guessNumByExample/static"
	staticHandler := http.FileServer(http.Dir(dir))
	staticHandler.ServeHTTP(w, req)
}

func main() {
	flag.Parse()

	logrus.SetLevel(logrus.DebugLevel)

	cc := blockchain.NewChainCode()

	hub := newHub()
	go hub.run()
	//http.HandleFunc("/", serveHome)
	http.HandleFunc("/", StaticServer)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r, cc)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
