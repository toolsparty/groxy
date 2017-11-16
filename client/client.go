package main

import (
	"log"
	"net/http"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/toolsparty/groxy/conf"
	"github.com/toolsparty/groxy/handlers"
	logger2 "github.com/toolsparty/groxy/logger"
)

var cfg = conf.NewConfiguration()
var logger = logger2.NewLogger("client", cfg)

func main() {
	defer logger.Close()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logger.Close()
		os.Exit(1)
	}()

	// for http
	go func() {
		logger.Write("Start listen", cfg.GetAddr())
		handler, _ := handlers.NewClient(cfg, logger, false)
		log.Fatal(http.ListenAndServe(cfg.GetAddr(), handler))
	}()

	// for https
	go func() {
		logger.Write("Start listen", cfg.GetAddrs())
		handler, _ := handlers.NewClient(cfg, logger,true)
		log.Fatal(http.ListenAndServe(cfg.GetAddrs(), handler))
	}()

	// wait
	var input string
	fmt.Scanln(&input)
	fmt.Println("done")
}
