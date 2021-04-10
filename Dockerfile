FROM golang:1.16-buster

RUN apt-get update && apt-get install -y --no-install-recommends \
		git \
	&& rm -rf /var/lib/apt/lists/*

ADD . /app

WORKDIR /app

RUN make test

CMD
