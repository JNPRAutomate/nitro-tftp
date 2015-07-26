package main

import (
	"log"
	"os"
	"os/signal"
)

func main() {
	s := &TFTPServer{}
	s.LoadConfig(&Config{})
	ctrlChan := s.Listen()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	close(c)
	close(ctrlChan)
	log.Println("Got", sig)
}
