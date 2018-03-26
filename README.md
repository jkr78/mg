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


List of stuff what is not done:
* Only one game per server
* Not tested carefully
* 512 bytes per line (client will be kicked out if sends long lines)


TODO:
* Interfaces
* As is now it is possible to extend transport side (again - interfaces)
* Do it in Golang style
* Fix grammar mistakes in documntation
* Add doc strings, document

