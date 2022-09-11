package main

import (
	"fmt"
	"github.com/basterrus/golang_backend_1/server/internal"
	"github.com/spf13/viper"
	"log"
	"net"
)

func main() {
	// Initialize configs files
	if err := internal.InitConfig(); err != nil {
		log.Fatalf("Error load configuration file %s", err.Error())
	}
	address := fmt.Sprintf(viper.GetString("ADDRESS") + ":" + viper.GetString("PORT"))
	network := viper.GetString("NETWORK")

	listener, err := net.Listen(network, address)
	if err != nil {
		log.Fatal(err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatal("Error close connection: ", err)
		}
	}(listener)

	go internal.Broadcaster()

	for {
		userSocket, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go internal.HandlerConn(userSocket)
	}
}
