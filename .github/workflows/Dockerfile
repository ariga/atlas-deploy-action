FROM arigaio/atlas:latest-alpine as atlas

FROM golang:1.20 as build

WORKDIR /go/src/app
COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian11

COPY --from=atlas /atlas .

COPY --from=build /go/bin/app /

CMD ["/app"]