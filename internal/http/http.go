package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var content string

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Invalid usage")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(0)
	}()

	switch os.Args[1] {
	case "client":
		resp, err := http.Get(os.Args[2])
		if err != nil {
			fmt.Printf("Failed: %s\n", err)
		}

		io.Copy(os.Stdout, resp.Body)
	case "server":
		content = os.Args[2]
		http.HandleFunc("/", Server)
		http.ListenAndServe(":8080", nil)
	}
}

func Server(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(content))
}
