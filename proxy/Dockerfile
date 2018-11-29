# STEP 1 build executable binary
FROM golang:alpine
RUN apk update && apk add git
RUN mkdir /go/src/app
COPY . /go/src/app
WORKDIR /go/src/app
RUN go get -d ./...
#build the binary
RUN go build -o proxy .
CMD ["./proxy"]





