package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

var StreamsAPIEndpointV5 = "https://api.twitch.tv/kraken/streams?limit=100"

type StreamsConsumer struct {
	Endpoint        string
	HTTPMethod      string
	TwitchToken     string
	RequestInterval int
	Parameters      map[string]string
	PublishChan     chan PublishEvent
	shutdown        bool
}

type StreamsConsumerResponse struct {
	Total   int              `json:"_total"`
	Streams []StreamResponse `json:"streams"`
}

type StreamResponse struct {
	ID          int               `json:"_id"`
	AverageFPS  float64           `json:"average_fps"`
	Channel     StreamChannel     `json:"channel"`
	CreatedAt   string            `json:"created_at"`
	Delay       int               `json:"delay"`
	Game        string            `json:"game"`
	IsPlaylist  bool              `json:"is_playlist"`
	Preview     map[string]string `json:"preview"`
	VideoHeight int               `json:"video_height"`
	Viewers     int               `json:"viewers"`
}

type StreamChannel struct {
	ID                int    `json:"_id"`
	BroadcastLanguage string `json:"broadcast_language"`
	CreatedAt         string `json:"created_at"`
	DisplayName       string `json:"display_name"`
	Followers         int    `json:"followers"`
	Game              string `json:"game"`
	Language          string `json:"language"`
	Logo              string `json:"logo"`
	Mature            bool   `json:"mature"`
	Name              string `json:"name"`
	Partner           bool   `json:"partner"`
	ProfileBanner     string `json:"profile_banner"`
	Status            string `json:"status"`
	UpdatedAt         string `json:"updated_at"`
	URL               string `json:"url"`
	VideoBanner       string `json:"video_banner"`
	Views             int    `json:"views"`
}

func (stream *StreamResponse) InfluxPoint() (map[string]string, map[string]interface{}) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["game"] = stream.Game
	tags["language"] = stream.Channel.Language
	tags["channel_id"] = strconv.Itoa(stream.Channel.ID)
	fields["created_at"] = stream.CreatedAt
	fields["name"] = stream.Channel.Name
	fields["status"] = stream.Channel.Status
	fields["viewers"] = stream.Viewers
	fields["channel_views"] = stream.Channel.Views

	return tags, fields
}

func DefaultStreamsConsumer(pubChan chan PublishEvent, config *Config) *StreamsConsumer {
	return &StreamsConsumer{
		Endpoint:        StreamsAPIEndpointV5,
		HTTPMethod:      "GET",
		TwitchToken:     config.TwitchAPIToken,
		RequestInterval: 300,
		Parameters:      map[string]string{},
		PublishChan:     pubChan,
		shutdown:        false,
	}
}

func (sc StreamsConsumer) Consume() {
	for !sc.shutdown {
		log.Println("StreamsConsumer making request...")

		fields := make(map[string]interface{})
		tags := map[string]string{"consumer_type": "streams_consumer"}
		req, err := http.NewRequest("GET", sc.Endpoint, nil)
		if err != nil {
			log.Println("Error creating Streams request:", err.Error())
		}

		req.Header.Add("accept", "application/vnd.twitchtv.v5+json")
		req.Header.Add("client-id", sc.TwitchToken)

		start := time.Now()
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("Error making clips request:", err.Error())
		}
		fields["response_time"] = time.Since(start)
		fields["reponse_code"] = res.StatusCode

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println("Error reading clips response:", err.Error())
		}

		var streamResponse StreamsConsumerResponse
		err = json.Unmarshal(body, &streamResponse)
		if err != nil {
			log.Println("Error unmarshalling data into StreamsConsumerResponse:", err.Error())
		} else {
			sc.PushStreamsToChannel(streamResponse.Streams)
		}

		res.Body.Close()
		sendStatsToInflux("consumers", tags, fields)
		time.Sleep(time.Duration(sc.RequestInterval) * time.Second)
	}

	log.Println("Clip Consumer shutting down...")
}

func (sc StreamsConsumer) Shutdown() {
	sc.shutdown = true
}

func (sc StreamsConsumer) PushStreamsToChannel(streams []StreamResponse) {
	for _, stream := range streams {
		event := PublishEvent{
			Type: "stream",
			Data: stream,
		}

		if !noinflux {
			tags, fields := stream.InfluxPoint()
			sendStatsToInflux("streams_consumer", tags, fields)
		}

		sc.PublishChan <- event
	}
}
