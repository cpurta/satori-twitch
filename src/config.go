package main

import (
	"fmt"
	"os"
)

type Config struct {
	TwitchAPIToken    string `json:"twitch_token"`
	SatoriChannelName string `json:"satori_channel_name"`
	SatoriEndpoint    string `json:"satori_endpoint"`
	SatoriAppKey      string `json:"satori_app_key"`
	SatoriRole        string `json:"satori_role"`
	SatoriSecret      string `json:"satori_secret"`
}

func LoadConfig() (*Config, error) {
	var err error
	twitchToken, err := GetRequiredEnvironmentVariable("TWITCH_TOKEN")
	if err != nil {
		return nil, err
	}
	satoriChannel, err := GetRequiredEnvironmentVariable("SATORI_CHANNEL")
	if err != nil {
		return nil, err
	}
	satoriEndpoint, err := GetRequiredEnvironmentVariable("SATORI_ENDPOINT")
	if err != nil {
		return nil, err
	}
	satoriAppKey, err := GetRequiredEnvironmentVariable("SATORI_APP_KEY")
	if err != nil {
		return nil, err
	}
	satoriRole, err := GetRequiredEnvironmentVariable("SATORI_ROLE")
	if err != nil {
		return nil, err
	}
	satoriSecret, err := GetRequiredEnvironmentVariable("SATORI_SECRET")
	if err != nil {
		return nil, err
	}

	conf := &Config{
		TwitchAPIToken:    twitchToken,
		SatoriChannelName: satoriChannel,
		SatoriEndpoint:    satoriEndpoint,
		SatoriAppKey:      satoriAppKey,
		SatoriRole:        satoriRole,
		SatoriSecret:      satoriSecret,
	}

	return conf, nil
}

func GetRequiredEnvironmentVariable(varName string) (string, error) {
	var v string
	v = os.Getenv(varName)
	if v == "" {
		return v, fmt.Errorf("Unset or empty variable (%s) provided as required env variable", varName)
	}

	return v, nil
}

func GetEnvironmentVariable(key string) string {
	return os.Getenv(key)
}
