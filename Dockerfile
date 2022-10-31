FROM golang:1.17-alpine

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
#COPY go.mod go.sum ./
#RUN go mod download && go mod verify
COPY go.mod ./
COPY *.go ./

COPY . .
RUN go build -v -ldflags "-w -s" -o /usr/local/bin/app ./...

ENTRYPOINT ["app"]
