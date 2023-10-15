FROM arigaio/atlas:latest-alpine as atlas

FROM golang:1.21 as build

WORKDIR /go/src/app
COPY . .

RUN go mod download

RUN go build -o /go/bin/app

COPY --from=atlas /atlas /bin

CMD ["/go/bin/app"]