package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/toolsparty/groxy/handlers"
	conf2 "github.com/toolsparty/groxy/conf"
	logger2 "github.com/toolsparty/groxy/logger"
)

var conf = conf2.NewConfiguration()
var logg = logger2.NewLogger("server", conf)

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
