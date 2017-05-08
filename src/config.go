package main

import (
	"fmt"
	"os"
)

type Config struct {
	InfluxAddr        string `json:"influx_address"`
	InfluxPort        string `json:"influx_port"`
	InfluxUsername    string `json:"influx_username"`
	InfluxPassword    string `json:"influx_password"`
	InfluxDatabase    string `json:"influx_database"`
	TwitchAPIToken    string `json:"twitch_token"`
	SatoriChannelName string `json:"satori_channel_name"`
	SatoriEndpoint    string `json:"satori_endpoint"`
	SatoriAppKey      string `json:"satori_app_key"`
	SatoriRole        string `json:"satori_role"`
	SatoriSecret      string `json:"satori_secret"`
}

func LoadConfig() (*Config, error) {
	var err error
	influxAddr := GetEnvironmentVariable("INFLUXDB_PORT_8086_TCP_ADDR")
	influxPort := GetEnvironmentVariable("INFLUXDB_PORT_8086_TCP_PORT")
	influxUsername := GetEnvironmentVariable("INFLUXDB_USERNAME")
	influxPassword := GetEnvironmentVariable("INFLUXDB_PASSWORD")
	influxDatabase := GetEnvironmentVariable("INFLUXDB_DATABASE")
	twitchToken, err := GetRequiredEnvironmentVariable("TWITCH_TOKEN")
	satoriChannel, err := GetRequiredEnvironmentVariable("SATORI_CHANNEL")
	satoriEndpoint, err := GetRequiredEnvironmentVariable("SATORI_ENDPOINT")
	satoriAppKey, err := GetRequiredEnvironmentVariable("SATORI_APP_KEY")
	satoriRole, err := GetRequiredEnvironmentVariable("SATORI_ROLE")
	satoriSecret, err := GetRequiredEnvironmentVariable("SATORI_SECRET")

	if err != nil {
		return nil, err
	}

	conf := &Config{
		InfluxAddr:        influxAddr,
		InfluxPort:        influxPort,
		InfluxUsername:    influxUsername,
		InfluxPassword:    influxPassword,
		InfluxDatabase:    influxDatabase,
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
