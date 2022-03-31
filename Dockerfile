# to build this docker image:
#   docker build .
FROM gocv/opencv:4.5.5

RUN apt-get update && apt-get install -y ffmpeg

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN go build -v -o video-color-palette-generator

RUN cp ./video-color-palette-generator /usr/local/bin

RUN video-color-palette-generator --help

ENTRYPOINT [ "video-color-palette-generator", "lambda"] 
