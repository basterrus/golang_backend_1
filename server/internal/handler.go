package internal

import (
	"bufio"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net"
	"path/filepath"
)

type client chan<- string

type message struct {
	userMessage string
	user        client
}

type who struct {
	addr     string
	nickname string
}

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan message)
)

func Broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				if msg.user != cli {
					cli <- msg.userMessage
				}
			}
		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

// InitConfig Load configs files
func InitConfig() error {
	path := filepath.Dir("configs/config.yml")
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

func HandlerConn(conn net.Conn) {
	ch := make(chan string)

	go clientWriter(conn, ch)

	fmt.Fprintln(conn, "Please, enter your username")

	var username string
	fmt.Fscanln(conn, &username)
	whoIs := who{addr: conn.RemoteAddr().String(), nickname: username}

	ch <- "You connected server at: " + whoIs.nickname + " (" + whoIs.addr + ")"
	messages <- message{userMessage: whoIs.nickname + " connected to server"}
	entering <- ch
	log.Println(whoIs.nickname + " connected to server")

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- message{userMessage: whoIs.nickname + ": " + input.Text(), user: ch}
	}

	leaving <- ch
	messages <- message{userMessage: whoIs.nickname + " disconnect"}
	log.Println(whoIs.nickname + " disconnect")
	err := conn.Close()
	if err != nil {
		return
	}
}

func clientWriter(c net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(c, msg)
	}
}
