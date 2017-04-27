package main

import "encoding/json"

type PublishEvent struct {
	Type string      `json:"event_name"`
	Data interface{} `json:"data"`
}

func (e *PublishEvent) MarshalString() (string, error) {
	body, err := json.Marshal(e)
	return string(body), err
}
