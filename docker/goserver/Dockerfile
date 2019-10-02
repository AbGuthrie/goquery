FROM golang:latest
LABEL maintainer="obelisk"

ENV GO113MODULE=on

RUN mkdir -p /goserver/certs
WORKDIR /goserver

COPY docker/certs/ certs/

COPY go.mod .
COPY go.sum .

# Copy the goserver code
COPY goserver/ goserver/

RUN go build -o bin/mock_osquery_server goserver/*.go

ENTRYPOINT [ "bin/mock_osquery_server" ]
CMD [ "-server_cert=/goserver/certs/example_server.crt", "-server_key=/goserver/certs/example_server.key" ]
# ENTRYPOINT [ "/bin/bash" ]
