FROM golang:alpine As build

WORKDIR /go/src/app

COPY go.mod ./
COPY go.sum ./
COPY ./ff ./ff

RUN go mod download

COPY *.go ./

RUN go build -o /ff-webgoapi

## Deploy
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY ./json ./json
COPY ./scripts ./scripts

COPY --from=build /ff-webgoapi /ff-webgoapi

EXPOSE 80

USER nonroot:nonroot

ENTRYPOINT [ "/ff-webgoapi" ]