package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var VideosAPIEndpointV5 = "https://api.twitch.tv/kraken/videos/top?period=month&limit=100&sort=views"

type VideosConsumer struct {
	Endpoint        string
	HTTPMethod      string
	TwitchToken     string
	RequestInterval int
	Parameters      map[string]string
	PublishChan     chan PublishEvent
	shutdown        bool
}

type VideosConsumerResponse struct {
	VODs []VideoResponse `json:"vods"`
}

type VideoResponse struct {
	ID              string       `json:"_id"`
	BroadcastID     int          `json:"broadcast_id"`
	BroadcastType   string       `json:"broadcast_type"`
	Channel         VideoChannel `json:"channel"`
	CreatedAt       string       `json:"created_at"`
	Description     string       `json:"description"`
	DescriptionHTML string       `json:"description_html"`
	Game            string       `json:"game"`
	Language        string       `json:"language"`
	Length          int          `json:"length"`
	PublishedAt     string       `json:"published_at"`
	Status          string       `json:"status"`
	TagList         string       `json:"tag_list"`
	Title           string       `json:"title"`
	URL             string       `json:"url"`
	Viewable        string       `json:"viewable"`
	Views           int          `json:"views"`
}

type VideoChannel struct {
	ID          float64 `json:"_id"`
	DisplayName string  `json:"display_name"`
	Name        string  `json:"name"`
}

func (video *VideoResponse) InfluxPoint() (map[string]string, map[string]interface{}) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["game"] = video.Game
	tags["language"] = video.Language
	fields["views"] = video.Views
	fields["title"] = video.Title
	fields["status"] = video.Status

	return tags, fields
}

func DefaultVideosConsumer(pubChan chan PublishEvent, config *Config) *VideosConsumer {
	return &VideosConsumer{
		Endpoint:        VideosAPIEndpointV5,
		HTTPMethod:      "GET",
		TwitchToken:     config.TwitchAPIToken,
		RequestInterval: 300,
		Parameters:      map[string]string{},
		PublishChan:     pubChan,
		shutdown:        false,
	}
}

func (vc VideosConsumer) Consume() {
	for !vc.shutdown {
		log.Println("VideosConsumer making request...")
		req, err := http.NewRequest("GET", vc.Endpoint, nil)
		if err != nil {
			log.Println("Error creating Videos request:", err.Error())
		}

		req.Header.Add("accept", "application/vnd.twitchtv.v5+json")
		req.Header.Add("client-id", vc.TwitchToken)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("Error making clips request:", err.Error())
			continue
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println("Error reading clips response:", err.Error())
			continue
		}

		var videoResponse VideosConsumerResponse
		err = json.Unmarshal(body, &videoResponse)
		if err != nil {
			log.Println("Error unmarshalling data into VideosConsumerResponse:", err.Error())
		} else {
			vc.PushStreamsToChannel(videoResponse.VODs)
		}

		res.Body.Close()
		time.Sleep(time.Duration(vc.RequestInterval) * time.Second)
	}

	log.Println("Clip Consumer shutting down...")
}

func (vc VideosConsumer) Shutdown() {
	vc.shutdown = true
}

func (vc VideosConsumer) PushStreamsToChannel(videos []VideoResponse) {
	for _, video := range videos {
		event := PublishEvent{
			Type: "video",
			Data: video,
		}

		vc.PublishChan <- event
	}
}
