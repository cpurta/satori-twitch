package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var ClipsAPIEndpointV5 = "https://api.twitch.tv/kraken/clips/top?limit=100"

type ClipsConsumer struct {
	Endpoint        string
	HTTPMethod      string
	TwitchToken     string
	RequestInterval int
	Parameters      map[string]string
	PublishChan     chan PublishEvent
	shutdown        bool
}

type ClipsConsumerResponse struct {
	Clips  []*ClipResponse `json:"clips"`
	Cursor string          `json:"cursor"`
}

type ClipResponse struct {
	Slug        string            `json:"slug"`
	TrackingID  string            `json:"tracking_id"`
	URL         string            `json:"url"`
	EmbedURL    string            `json:"embed_url"`
	EmbedHTML   string            `json:"embed_html"`
	Broadcaster map[string]string `json:"broadcaster"`
	Curator     map[string]string `json:"curator"`
	VOD         map[string]string `json:"vod"`
	Game        string            `json:"game"`
	Language    string            `json:"language"`
	Title       string            `json:"title"`
	Views       float64           `json:"views"`
	Duration    float64           `json:"duration"`
	CreatedAt   string            `json:"created_at"`
	Thumbnails  map[string]string `json:"thumbnails"`
}

func (clip *ClipResponse) InfluxPoint() (map[string]string, map[string]interface{}) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["clip_title"] = clip.Title
	tags["clip_game"] = clip.Game
	fields["tracking_id"] = clip.TrackingID
	fields["language"] = clip.Language
	fields["created_at"] = clip.CreatedAt
	fields["view"] = clip.Views

	return tags, fields
}

func DefaultClipsConsumer(pubChan chan PublishEvent, config *Config) *ClipsConsumer {
	return &ClipsConsumer{
		Endpoint:        ClipsAPIEndpointV5,
		HTTPMethod:      "GET",
		RequestInterval: 120,
		TwitchToken:     config.TwitchAPIToken,
		PublishChan:     pubChan,
		Parameters:      map[string]string{},
		shutdown:        false,
	}
}

func (cc ClipsConsumer) Consume() {
	for !cc.shutdown {
		log.Println("ClipsConsumer making request...")

		fields := make(map[string]interface{})
		tags := map[string]string{"consumer_type": "clips_consumer"}
		req, err := http.NewRequest("GET", cc.Endpoint, nil)
		if err != nil {
			log.Println("Error creating Clips request:", err.Error())
		}

		req.Header.Add("accept", "application/vnd.twitchtv.v5+json")
		req.Header.Add("client-id", cc.TwitchToken)

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

		var clips ClipsConsumerResponse
		err = json.Unmarshal(body, &clips)
		if err != nil {
			log.Println("Error unmarshalling data into ClipsConsumerResponse:", err.Error())
		} else {
			cc.PushClipsToChannel(clips)
		}

		res.Body.Close()
		sendStatsToInflux("consumers", tags, fields)
		time.Sleep(time.Duration(cc.RequestInterval) * time.Second)
	}

	log.Println("Clip Consumer shutting down...")
}

func (cc ClipsConsumer) Shutdown() {
	cc.shutdown = true
}

func (cc ClipsConsumer) PushClipsToChannel(clips ClipsConsumerResponse) {
	for _, clip := range clips.Clips {
		event := PublishEvent{
			Type: "clip",
			Data: clip,
		}

		if !noinflux {
			tags, fields := clip.InfluxPoint()
			sendStatsToInflux("clips_consumer", tags, fields)
		}

		cc.PublishChan <- event
	}
}
