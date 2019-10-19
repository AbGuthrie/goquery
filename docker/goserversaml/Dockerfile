FROM golang:latest
LABEL maintainer="obelisk"

ENV GO113MODULE=on

RUN mkdir -p /goserversaml/certs
WORKDIR /goserversaml

COPY docker/certs/ certs/

COPY go.mod .
COPY go.sum .

# Copy the goserver code
COPY goserversaml/ goserversaml/

RUN go build -o bin/goserversaml goserversaml/*.go

ENTRYPOINT [ "bin/goserversaml" ]
#CMD [ "-server_cert=/goserver/certs/example_server.crt", "-server_key=/goserver/certs/example_server.key" ]
# ENTRYPOINT [ "/bin/bash" ]
