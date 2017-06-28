package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var GamesAPIEndpointV5 = "https://api.twitch.tv/kraken/games/top?limit=100"

type GamesConsumer struct {
	Endpoint        string
	HTTPMethod      string
	TwitchToken     string
	RequestInterval int
	Parameters      map[string]string
	PublishChan     chan PublishEvent
	shutdown        bool
}

type GamesConsumerResponse struct {
	Total int            `json:"_total"`
	Top   []GameResponse `json:"top"`
}

type GameResponse struct {
	Channels int  `json:"channels"`
	Viewers  int  `json:"viewers"`
	Game     Game `json:"game"`
}

type Game struct {
	ID         int               `json:"_id"`
	Box        map[string]string `json:"box"`
	GiantID    int               `json:"giantbomb_id"`
	Logo       map[string]string `json:"logo"`
	Name       string            `json:"name"`
	Popularity int               `json:"popularity"`
}

func (game *GameResponse) InfluxPoint() (map[string]string, map[string]interface{}) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["name"] = game.Game.Name
	fields["popularity"] = game.Game.Popularity
	fields["viewers"] = game.Viewers

	return tags, fields
}

func DefaultGamesConsumer(pubChan chan PublishEvent, config *Config) *GamesConsumer {
	return &GamesConsumer{
		Endpoint:        GamesAPIEndpointV5,
		HTTPMethod:      "GET",
		TwitchToken:     config.TwitchAPIToken,
		RequestInterval: 300,
		Parameters:      map[string]string{},
		PublishChan:     pubChan,
		shutdown:        false,
	}
}

func (gc GamesConsumer) Consume() {
	for !gc.shutdown {
		log.Println("GamesConsumer making request...")

		req, err := http.NewRequest("GET", gc.Endpoint, nil)
		if err != nil {
			log.Println("Error creating Games request:", err.Error())
		}

		req.Header.Add("accept", "application/vnd.twitchtv.v5+json")
		req.Header.Add("client-id", gc.TwitchToken)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("Error making games request:", err.Error())
			continue
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println("Error reading games response:", err.Error())
			continue
		}

		var gameResponse GamesConsumerResponse
		err = json.Unmarshal(body, &gameResponse)
		if err != nil {
			log.Println("Error unmarshalling data into GamesConsumerResponse:", err.Error())
		} else {
			gc.PushGamesToChannel(gameResponse.Top)
		}

		res.Body.Close()
		time.Sleep(time.Duration(gc.RequestInterval) * time.Second)
	}

	log.Println("Clip Consumer shutting down...")
}

func (gc GamesConsumer) Shutdown() {
	gc.shutdown = true
}

func (gc GamesConsumer) PushGamesToChannel(games []GameResponse) {
	for _, game := range games {
		event := PublishEvent{
			Type: "game",
			Data: game,
		}

		gc.PublishChan <- event
	}
}
