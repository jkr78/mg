/* Game
 */
package mg

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	BoardWidth        = 10
	BoardHeight       = 20
	LoopSpin          = 2 // secs
	DefaultZombieNick = "night-king"
)

type Arrow struct {
	Shooter string
	X       int
	Y       int
}

type clientInfo struct {
	nick   string
	events chan string
	kills  int
}

type zombieInfo struct {
	x    int
	y    int
	nick string
}

type Game struct {
	width    int
	height   int
	speed    int
	loopStop chan bool
	loopDone chan bool
	shoots   chan Arrow
	clients  map[string]*clientInfo
	zombie   zombieInfo
	kills    int
	mux      sync.Mutex
	started  bool
}

func NewGame(width int, height int, speed int) *Game {
	game := Game{width: width, height: height, speed: speed}
	game.loopStop = make(chan bool)
	game.loopDone = make(chan bool)
	game.shoots = make(chan Arrow)
	game.clients = make(map[string]*clientInfo)

	rand.New(rand.NewSource(time.Now().UnixNano()))

	return &game
}

func (game *Game) AddClient(nick string, events chan string) error {
	game.mux.Lock()
	defer game.mux.Unlock()

	if _, ok := game.clients[nick]; ok {
		log.Printf("User is already registered: %s\n", nick)
		return errors.New(fmt.Sprintf("User with this nick is already registered"))
	}

	game.clients[nick] = &clientInfo{nick: nick, events: events}

	return nil
}

func (game *Game) RemoveClient(nick string) error {
	game.mux.Lock()
	defer game.mux.Unlock()

	delete(game.clients, nick)
	return nil
}

func (game *Game) Shoot(arrow Arrow) error {
	if arrow.X < 0 || arrow.X >= game.width {
		log.Printf("Bad X position from the client (%d, %d): %d\n",
			game.width, game.height, arrow.X)
		return errors.New(fmt.Sprintf("You broke the anti balistic missile treaty"))
	}

	if arrow.Y < 0 || arrow.Y >= game.height {
		log.Printf("Bad Y position from the client (%d, %d): %d\n",
			game.width, game.height, arrow.Y)
		return errors.New(fmt.Sprintf("Intercontinental missiles are banned"))
	}

	game.shoots <- arrow
	return nil
}

func (game *Game) Start() {
	game.mux.Lock()
	defer game.mux.Unlock()
	if game.started {
		return
	}

	go game.loop()
	game.started = true
	log.Println("Game has been started")
}

func (game *Game) Stop() {
	game.mux.Lock()
	defer game.mux.Unlock()

	if !game.started {
		return
	}

	// notify all clients
	for _, ci := range game.clients {
		ci.events <- "" // stop
	}

	game.loopStop <- true
	<-game.loopDone

	game.clients = make(map[string]*clientInfo)
	game.started = false
	log.Println("Game stopped")
}

// this function will block if client is not reading fast enought
// i could run it as coroutine but in the end it will still be a problem
// since it could create millions of coroutines if client is very slow
// queue/channel with round buffer is needed here, but i had no time
// to play around and find/make myself one
func (game *Game) sendEvent(msg string) {
	game.mux.Lock()
	defer game.mux.Unlock()

	for _, ci := range game.clients {
		ci.events <- msg
	}
}

func (game *Game) shootAndKill(ci *clientInfo) {
	ci.kills += 1
	game.sendEvent(fmt.Sprintf("BOOM %s %d %s",
		ci.nick, ci.kills, game.zombie.nick))

	game.killZombie()
}

func (game *Game) shootAndMiss(ci *clientInfo) {
	game.sendEvent(fmt.Sprintf("BOOM %s %d",
		ci.nick, ci.kills))
}

func (game *Game) killZombie() {
	log.Printf("Zombie killed: %s (%d, %d)",
		game.zombie.nick, game.zombie.x, game.zombie.y)
	game.zombie.nick = ""
}

func (game *Game) spawnZombie() {
	// here we could implement more random stuff
	game.zombie.x = 0
	game.zombie.y = 0
	game.zombie.nick = DefaultZombieNick
	log.Printf("New zombie spawned: %s (%d, %d)",
		game.zombie.nick, game.zombie.x, game.zombie.y)
}

func (game *Game) moveZombie() {
	if game.zombie.nick == "" && len(game.clients) > 0 { // new zombie
		game.spawnZombie()
	}

	if rand.Int31n(10)%3 == 0 { // vertical
		if rand.Int31n(10)%2 == 0 { // down
			if game.zombie.y < game.height-1 {
				game.zombie.y += 1
			} else {
				game.zombie.y -= 1
			}
		} else { // up
			if game.zombie.y > 0 {
				game.zombie.y -= 1
			} else {
				game.zombie.y += 1
			}
		}
	} else { // horizontal
		game.zombie.x += 1
	}

	game.sendEvent(fmt.Sprintf("WALK %s %d %d",
		game.zombie.nick, game.zombie.x, game.zombie.y))

	if game.zombie.x >= game.width {
		game.zombiesWins()
	}
}

func (game *Game) zombiesWins() {
	log.Printf("Westeros is doomed\n")
	game.zombie.nick = ""
}

func (game *Game) loop() {
	tick := time.NewTicker(time.Duration(game.speed) * time.Second).C

	for {
		select {
		case <-tick:
			game.moveZombie()
		case arrow := <-game.shoots:
			ci := game.clients[arrow.Shooter]
			if ci == nil {
				log.Printf("Got shoot command from unknown client: %s (%d, %d)",
					arrow.Shooter, arrow.X, arrow.Y)
				game.sendEvent("CHEATER!")
				break
			}
			if arrow.X == game.zombie.x && arrow.Y == game.zombie.y {
				game.shootAndKill(ci)
			} else {
				game.shootAndMiss(ci)
			}
		case <-game.loopStop:
			break
		}
	}

	game.loopDone <- true
}
