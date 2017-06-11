# satori-twitch

A simple project that polls the twitch API for videos, games, clips, and streams
and republishes that data to the Satori platform.

## Setup

The project uses wgo to build and manage the dependencies of the project. You can
initialize the project by running `wgo restore` in the project root.

## Building

You first need to make a `bin` directory by running `mkdir bin` in the project root.

You can build the binary by going into the `src` directory and running `wgo build -o ../bin/satori-twitch`

## Running

Before running the binary you will need to source the necessary environment variables
so that the program knows how to connect to Satori and have the authorization token to
access the Twitch API.

I recommend that you create a env file that looks like the following:

*example.env*
```
export SATORI_CHANNEL=<your_satori_channel>
export SATORI_ENDPOINT=<your_satori_endpoint>
export SATORI_APP_KEY=<your_satori_app_key>
export SATORI_ROLE=<your_satori_role>
export SATORI_SECRET=<your_satori_secret>
export TWITCH_TOKEN=<your_twitch_token>
```

You can then make this accessible to the program by sourcing the file:

```
$ source example.env
```

The program requires that it connects to an InfluxDB but you can force it to not
require to connect by specifying the `--no-influx` flag when running.

## Using docker-compose

Currently I use the program with docker-compose so that statistics can be monitored
by using InfluxDB and Grafana.

You will have to create a `docker-compose.yml` file similar the following in your project root:

```yaml
satori-twitch:
    build: .
    dockerfile: ./docker/Dockerfile
    links:
        - influxdb
    environment:
        INFLUXDB_USERNAME: "root"
        INFLUXDB_PASSWORD: "temppwd"
        INFLUXDB_DATABASE: "satori-twitch"
        TWITCH_TOKEN: "<your_twitch_token>"
        SATORI_CHANNEL: "<your_satori_channel>"
        SATORI_ENDPOINT: "<your_satori_endpoint>"
        SATORI_APP_KEY: "<your_satori_app_key>"
        SATORI_ROLE: "<your_satori_role>"
        SATORI_SECRET: "<your_satori_secret>"

influxdb:
    image: tutum/influxdb:latest
    ports:
        - 8086:8086
        - 8083:8083
    environment:
        - PRE_CREATE_DB="satori-twitch"
        - ADMIN_USER="root"
        - INFLUXDB_INIT_PWD="temppwd"

grafana:
    image: grafana/grafana:latest
    ports:
        - 3000:3000

```

Once you have the binary built and placed in the `bin` folder you can then run:

```
$ docker-compose build && docker-compose up
```
