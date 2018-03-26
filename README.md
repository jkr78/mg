# mg

Little game. Used as a Golang learning exercise.

Following features should be supported:
* Multi client
* Support for different line ending (per server)
* Configurable board width and height
* Configurable speed
* Client can join and close communication. Game loop runs
as long as where are one client started the game or zombie.
* When where is no more zombies/clients game loop "pauses".
* User can connect/disconnect in the middle of the game.
* Users with same nick are not allowed


How to test:
* `git clone https://github.com/jkr78/mg.git`
* `cd mg`
* `env GOPATH=$(pwd) go install github.com/jkr78/mg` - this will install shared library
* `env GOPATH=$(pwd) go run src/mgserv.go` - this will start game server
* in another terminal run `nc localhost 1201` and write `START me`


List of stuff what is not done:
* Only one game per server
* Not tested carefully
* 512 bytes per line (client will be kicked out if sends long lines)


TODO:
* Interfaces
* As is now it is possible to extend transport side (again - interfaces)
* Do it in Golang style
* Fix grammar mistakes in documntation
* Test and fix with https://goreportcard.com/
* Add Makefile to help build/install
* Add doc strings, document

