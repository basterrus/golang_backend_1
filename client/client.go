package main

import (
	"fmt"
	"github.com/spf13/viper"
	"io"
	"log"
	"net"
	"os"
)

func main() {

	// Initialize configs files
	if err := initConfig(); err != nil {
		log.Fatalf("Error load configuration file %s", err.Error())
	}
	address := fmt.Sprintf(viper.GetString("ADDRESS") + ":" + viper.GetString("PORT"))
	network := viper.GetString("NETWORK")

	conn, err := net.Dial(network, address)
	if err != nil {
		log.Fatal(err)
	}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(conn)

	go func() {
		_, err := io.Copy(os.Stdout, conn)
		if err != nil {
			return
		}
	}()

	io.Copy(conn, os.Stdin)
	fmt.Printf("%s: exit", conn.LocalAddr())
}

//Load configs files
func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
