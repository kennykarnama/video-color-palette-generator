package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/mccutchen/palettor"

	gim "github.com/ozankasikci/go-image-merge"
	"gocv.io/x/gocv"
)

type ColorPalette struct {
	Color  color.Color
	Weight float64
}

func main() {
	// parse args
	videoFilePath := os.Args[1]

	segmentDurationSeconds, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		panic(err)
	}

	outputFolder := os.Args[3]

	os.MkdirAll(outputFolder, os.ModePerm)

	paletteSize, _ := strconv.ParseInt(os.Args[4], 10, 32)

	iteration, _ := strconv.ParseInt(os.Args[5], 10, 32)

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
			// scale frame
			scaledVideoFrame := gocv.NewMat()
			gocv.Resize(videoFrame, &scaledVideoFrame, image.Point{X: 0, Y: 0}, 0.1, 0.1, gocv.InterpolationCubic)
			writeStatus := gocv.IMWriteWithParams(frameFileName, videoFrame, []int{gocv.IMWritePngStrategy})
			if !writeStatus {
				log.Fatalf("Error write file=%v", frameFileName)
			}
			// generate palette
			tmpImage, err := scaledVideoFrame.ToImage()
			if err != nil {
				log.Fatalf("Error convert Mat to image err=%v", err)
			}
			p, err := palettor.Extract(int(paletteSize), int(iteration), tmpImage)
			if err != nil {
				log.Fatalf("Palettor err=%v", err)
			}
			colorPalette := []*ColorPalette{}

			colors := p.Colors()

			sort.Slice(colors, func(i, j int) bool {
				return p.Weight(colors[i]) < p.Weight(colors[j])
			})

			for _, color := range colors {
				log.Printf("color: %v; weight: %v", color, p.Weight(color))
				colorPalette = append(colorPalette, &ColorPalette{
					Color:  color,
					Weight: p.Weight(color),
				})
			}

			paletteFileName := filepath.Join(outputFolder, fmt.Sprintf("palette_%v__segment_%v.png", frameCount, period+1))
			log.Printf("writing palette file=%v", paletteFileName)

			_, err = createPalette(paletteFileName, p.Colors())
			if err != nil {
				log.Fatalf("err=%v", err)
			}

			visualizeFileName := filepath.Join(outputFolder, fmt.Sprintf("visualize_%v__segment_%v.png", frameCount, period+1))
			// merge frame and palette to allow better visualization
			err = visualize(frameFileName, paletteFileName, visualizeFileName)
			if err != nil {
				log.Fatalf("err=%v", err)
			}

			scaledVideoFrame.Close()
		} else {
			frameCount = 0
			period++
		}
	}

}

func createPalette(out string, colors []color.Color) (image.Image, error) {
	blocks := len(colors)
	blockw := 1280 / blocks
	space := 0
	img := image.NewRGBA(image.Rect(0, 0, blocks*blockw+space*(blocks-1), 4*(blockw+space)))

	for i := 0; i < blocks; i++ {
		draw.Draw(img, image.Rect(i*(blockw+space), 0, i*(blockw+space)+blockw, blockw), &image.Uniform{colors[i]}, image.Point{}, draw.Src)
	}

	toimg, err := os.Create(out)
	if err != nil {
		return nil, fmt.Errorf("action=createPalette out=%v err=%v", out, err)
	}
	defer toimg.Close()

	err = png.Encode(toimg, img)
	if err != nil {
		return nil, fmt.Errorf("action=createPalette out=%v err=%v", out, err)
	}

	return img, nil
}

func visualize(videoFramePath string, palettePath string, out string) error {
	grids := []*gim.Grid{
		{ImageFilePath: videoFramePath},
		{ImageFilePath: palettePath},
	}
	rgba, err := gim.New(grids, 1, 2).Merge()
	if err != nil {
		return fmt.Errorf("action=visualize video_frame_path=%v palette_path=%v err=%v", videoFramePath, palettePath, err)
	}
	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("action=visualize video_frame_path=%v palette_path=%v err=%v", videoFramePath, palettePath, err)
	}
	err = png.Encode(f, rgba)
	if err != nil {
		return fmt.Errorf("action=visualize video_frame_path=%v palette_path=%v err=%v", videoFramePath, palettePath, err)
	}
	return nil
}
