/* Client
 */
package mg

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	MaxBufSize   = 512
	ReadTimout   = 5 * 60
	WriteTimeout = 5
)

var LF = [...]byte{'\n'}
var CRLF = [...]byte{'\r', '\n'}

type Client struct {
	sep      []byte
	game     *Game
	conn     net.Conn
	nick     string
	events   chan string
	commands chan string
	stopGame chan bool
	gameDone chan bool
}

func NewClient(game *Game, conn net.Conn, sep []byte) *Client {
	client := Client{sep: sep, game: game, conn: conn}
	client.events = make(chan string)
	client.commands = make(chan string)
	client.stopGame = make(chan bool)
	client.gameDone = make(chan bool)
	return &client
}

func (client *Client) Play() {
	go client.socketHandler()

	for {
		select {
		case event := <-client.events:
			if event == "" {
				// killed from server
				break
			} else {
				client.send(event)
			}
		case command := <-client.commands:
			if err := client.parseCommand(command); err != nil {
				client.sendError(err.Error())
			}
		case <-client.stopGame:
			client.LeaveGame()
			break
		}
	}

	client.conn.Close()
	client.gameDone <- true
}

func (client *Client) JoinGame(nick string) error {
	err := client.game.AddClient(nick, client.events)
	if err != nil {
		log.Println("Cannot add the client: %s\n", err)
		return errors.New(fmt.Sprintf("Cannot add the client: %s", err))
	}

	client.nick = nick
	return nil
}

func (client *Client) LeaveGame() error {
	client.game.RemoveClient(client.nick)

	client.stopGame <- true
	<-client.gameDone

	client.nick = ""
	return nil
}

func (client *Client) Shoot(x, y int) error {
	arrow := Arrow{Shooter: client.nick, X: x, Y: y}
	log.Printf("Client is shooting the arrow: %d %d (%s)\n",
		arrow.X, arrow.Y, arrow.Shooter)
	client.game.Shoot(arrow)
	return nil
}

func (client *Client) send(msg string) {
	client.conn.Write(bytes.Join([][]byte{[]byte(msg), client.sep[:]}, []byte("")))
}

func (client *Client) sendError(err string) {
	// my own extension of the protocol
	client.send(fmt.Sprintf("ERROR \"%s\"", err))
}

// copy from https://gobyexample.com/collection-functions
func filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func (client *Client) parseCommand(command string) error {
	tokens := filter(strings.Split(command, " "), func(s string) bool {
		return s != ""
	})

	log.Printf("tokens: %q", tokens)
	if len(tokens) < 1 {
		return nil
	}

	onStartCommand := func(args []string) error {
		// do not know if i should be very strict or just ingnore extra args
		if len(args) != 1 {
			return errors.New(fmt.Sprintf("Bad number of arguments"))
		}

		fmt.Printf("Starting client: %s\n", args[0])
		if err := client.JoinGame(args[0]); err != nil {
			return err
		}

		client.game.Start()
		return nil
	}

	onShootCommand := func(args []string) error {
		if len(args) != 2 {
			return errors.New(fmt.Sprintf("Bad number of arguments"))
		}

		var x, y int
		var err error

		x, err = strconv.Atoi(args[0])
		if err != nil {
			return errors.New(fmt.Sprintf("Bad X position: %s", x))
		}
		y, err = strconv.Atoi(args[1])
		if err != nil {
			return errors.New(fmt.Sprintf("Bad Y position: %s", y))
		}

		return client.Shoot(x, y)
	}

	switch strings.ToUpper(tokens[0]) {
	case "": // nop command
		// do nothing
		return nil
	case "START":
		return onStartCommand(tokens[1:])
	case "SHOOT":
		return onShootCommand(tokens[1:])
	default:
		return errors.New(fmt.Sprintf("Unknown command: %s", tokens[0]))
	}
}

func (client *Client) socketHandler() {
	defer client.conn.Close()
	defer client.LeaveGame()

	var buf = make([]byte, MaxBufSize)
	bufLen := 0
	for {
		if MaxBufSize-bufLen <= 0 {
			log.Printf("Not enought space to receive command: (%d, %d)\n", bufLen-MaxBufSize, bufLen)
			return
		}

		n, err := client.conn.Read(buf[bufLen:])
		if err != nil {
			log.Printf("Error while reading data: %s\n", err)
			return
		}

		if n == 0 {
			// connection reset by peer
			log.Println("Connection reset by peer")
			return
		}

		bufLen += n
		for {
			i := bytes.IndexByte(buf, client.sep[len(client.sep)-1])
			if i < len(client.sep)-1 {
				break
			}
			found := true
			for j := 1; j < len(client.sep); j++ {
				if buf[i-j] != client.sep[len(client.sep)-j-1] {
					found = false
					break
				}
			}
			if !found {
				break
			}

			// lets support UTF
			if i > 0 {
				client.commands <- string(buf[:i-len(client.sep)+1])
			}

			buf = buf[i+1:]
			bufLen -= (i + 1)
		}
	}
}
