package main

type TwitchAPIConsumer interface {
	Consume()
	Shutdown()
}
