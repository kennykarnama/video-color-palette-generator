package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"gocv.io/x/gocv"
)

func main() {
	videoFilePath := os.Args[1]

	segmentDurationSeconds, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		panic(err)
	}

	outputFolder := os.Args[3]

	os.MkdirAll(outputFolder, os.ModePerm)

	vc, err := gocv.VideoCaptureFile(videoFilePath)
	if err != nil {
		panic(err)
	}
	defer vc.Close()

	videoFps := vc.Get(gocv.VideoCaptureFPS)
	frameCountsPerSegment := math.Ceil(videoFps * segmentDurationSeconds)

	log.Printf("FPS=%v", videoFps)
	log.Printf("FrameCountsPerSegment=%v", frameCountsPerSegment)

	videoFrame := gocv.NewMat()
	defer videoFrame.Close()

	frameCount := float64(0)
	period := 0

	for {
		if ok := vc.Read(&videoFrame); !ok {
			log.Printf("Video frame closed")
			return
		}
		if videoFrame.Empty() {
			log.Printf("Video frame empty")
			continue
		}

		frameCount++

		if frameCount <= frameCountsPerSegment {
			frameFileName := filepath.Join(outputFolder, fmt.Sprintf("frame_%v__segment_%v.png", frameCount, period+1))
			log.Printf("writing file=%v", frameFileName)
			writeStatus := gocv.IMWriteWithParams(frameFileName, videoFrame, []int{gocv.IMWritePngStrategy})
			if !writeStatus {
				log.Fatalf("Error write file=%v", frameFileName)
			}
		} else {
			frameCount = 0
			period++
		}
	}

}
