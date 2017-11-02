package main

import (
	"config"
	"handlers"
	"log"
	"net/http"
	logger3 "logger"
	"os"
	"os/signal"
	"syscall"
)

var conf = config.NewConfiguration()
var logg = logger3.NewLogger("server", conf)

func main() {
	defer logg.Close()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logg.Close()
		os.Exit(1)
	}()

	handler, _ := handlers.NewServer(conf, logg)
	log.Fatal(http.ListenAndServe(conf.Server.GetAddr(), handler))
}
