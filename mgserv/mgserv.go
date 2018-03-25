/* Server
 */
package main

import (
	"flag"
	"fmt"
	"github.com/jkr78/mg"
	"net"
	"os"
	"strconv"
)

func main() {
	host := flag.String("host", "", "host")
	port := flag.Int("port", 1201, "port")
	width := flag.Int("width", mg.BoardWidth, "board width")
	height := flag.Int("height", mg.BoardWidth, "board height")
	speed := flag.Int("speed", -mg.LoopSpin, "speed")
	crlf := flag.Bool("crlf", false, "use CRLF as end of line (default LF)")

	flag.Parse()

	addr := net.JoinHostPort(*host, strconv.Itoa(*port))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
		os.Exit(1)
	}

	var sep []byte
	if *crlf {
		fmt.Println("CRLF")
		sep = mg.CRLF[:]
	} else {
		fmt.Println("LF")
		sep = mg.LF[:]
	}

	game := mg.NewGame(*width, *height, -*speed)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		client := mg.NewClient(game, conn, sep)
		go client.Play()
	}
}
