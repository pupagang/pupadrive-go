FROM golang:rc-buster

WORKDIR /go/src/app
COPY . .

# install and update system dependencies 
RUN apt update && apt upgrade -y && apt-get install -y build-essential 

# install dependencies
RUN go get -d -v ./...

# build
RUN CGO_ENABLED=0 go build cmd/pupadrive/main.go

# cleanup uneccessary files & cache
RUN rm -rf go.* cmd internal *.md *.go pkg && go clean

RUN mkdir downloads

CMD ["./main"]