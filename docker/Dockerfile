FROM ubuntu:latest

MAINTAINER Chris Purta cpurta@gmail.com

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    mkdir -p /opt

ADD ./bin/satori-twitch /opt

RUN chmod +x /opt/satori-twitch

ENTRYPOINT ["/opt/satori-twitch"]
