# STEP 1 build executable binary
FROM golang:alpine as builder
RUN apk update && apk add git
RUN mkdir /go/src/app
COPY . /go/src/app
WORKDIR /go/src/app
RUN go get -d ./...
#build the binary
RUN go build -o graphserver .

# STEP 2 build a small image
FROM alpine
RUN apk update && apk add git && apk add ca-certificates
COPY --from=builder /go/src/app/graphserver /
CMD ["/graphserver"]




