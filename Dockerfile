FROM golang:1.20

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

USER burp
COPY . .
RUN go build -v -o /usr/local/bin cmd/burp-agent/app.go

EXPOSE 8873
CMD ["app"]