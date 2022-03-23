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

	"gitlab.com/ruangguru/kennykarnama/video-color-palette-generator/ffprobe"

	"github.com/gocarina/gocsv"

	"github.com/mccutchen/palettor"

	"time"

	"github.com/alexflint/go-arg"
	gim "github.com/ozankasikci/go-image-merge"
	"github.com/satori/go.uuid"
	"gocv.io/x/gocv"
)

type Result struct {
	SourceSerial          string  `csv:"source_serial"`
	SourceURL             string  `csv:"source_url"`
	SourceDurationSeconds float64 `csv:"source_duration_seconds"`
	SourceFPS             float64 `csv:"source_fps"`
	SampleID              string  `csv:"sample_id"`
	SampleNumber          int     `csv:"sample_number"`
	SampleDuration        float64 `csv:"sample_duration"`
	PaletteID             string  `csv:"palette_id"`
	PaletteCounts         int     `csv:"palette_counts"`
	R                     uint32  `csv:"r"`
	G                     uint32  `csv:"g"`
	B                     uint32  `csv:"b"`
	A                     uint32  `csv:"a"`
	RNorm                 float64 `csv:"r_norm"`
	GNorm                 float64 `csv:"g_norm"`
	BNorm                 float64 `csv:"b_norm"`
	Weight                float64 `csv:"weight"`
}

func (r *Result) Normalize16BitRGB() {
	r.RNorm = float64(r.R) / 65535.0
	r.GNorm = float64(r.G) / 65535.0
	r.BNorm = float64(r.B) / 65535.0
}

var args struct {
	InputFile      string         `arg:"required,--input-file,-i" help:"input file path for video"`
	PeriodDuration float64        `arg:"required,--period-duration,-d" help:"period duration in seconds"`
	PaletteSize    int            `arg:"required,--palette-size,-k" help:"palette size"`
	MaxIteration   int            `arg:"--max-iteration" default:"300" help:"maximum iteration of k-means if not convergent"`
	CsvResult      string         `arg:"required,--csv-result,-o" help:"csv result path"`
	VisualizeCmd   *VisualizeArgs `arg:"subcommand:visualize"`
}

type VisualizeArgs struct {
	OutputFolder string `arg:"--visualize-output-folder" help:"visualization output folder. Contains frame and color palette"`
}

func main() {
	// parse args
	arg.MustParse(&args)

	videoFilePath := args.InputFile

	segmentDurationSeconds := args.PeriodDuration

	paletteSize := args.PaletteSize

	iteration := args.MaxIteration

	resultFilePath := args.CsvResult

	outputFolder := ""

	if args.VisualizeCmd != nil {
		outputFolder = args.VisualizeCmd.OutputFolder

		os.MkdirAll(outputFolder, os.ModePerm)
	}

	vc, err := gocv.VideoCaptureFile(videoFilePath)
	if err != nil {
		panic(err)
	}
	defer vc.Close()

	prober, err := ffprobe.NewFfprobe(videoFilePath)
	if err != nil {
		panic(err)
	}

	videoFps := prober.GetVideoFps()
	videoDuration := prober.GetDuration().Seconds()
	frameCountsPerSegment := math.Ceil(videoFps * segmentDurationSeconds)

	log.Printf("FPS=%v", videoFps)
	log.Printf("FrameCountsPerSegment=%v", frameCountsPerSegment)
	log.Printf("durationSeconds=%v", videoDuration)

	videoFrame := gocv.NewMat()
	defer videoFrame.Close()

	frameCount := float64(0)
	period := 0
	periodID := uuid.NewV4().String()

	var results []*Result

	f, err := os.OpenFile(resultFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer f.Close()

	defer timeTrack(time.Now(), "video-color-palette-extraction")

	start := time.Now()

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
			// scale frame
			scaledVideoFrame := gocv.NewMat()
			gocv.Resize(videoFrame, &scaledVideoFrame, image.Point{X: 0, Y: 0}, 0.1, 0.1, gocv.InterpolationCubic)

			// generate palette
			tmpImage, err := scaledVideoFrame.ToImage()
			if err != nil {
				log.Fatalf("Error convert Mat to image err=%v", err)
			}
			p, err := palettor.Extract(int(paletteSize), int(iteration), tmpImage)
			if err != nil {
				log.Fatalf("Palettor err=%v", err)
			}

			colors := p.Colors()

			sort.Slice(colors, func(i, j int) bool {
				return p.Weight(colors[i]) < p.Weight(colors[j])
			})

			if args.VisualizeCmd != nil {
				frameFileName := filepath.Join(outputFolder, fmt.Sprintf("frame_%v__segment_%v.png", frameCount, period+1))
				log.Printf("writing file=%v", frameFileName)
				writeStatus := gocv.IMWriteWithParams(frameFileName, videoFrame, []int{gocv.IMWritePngStrategy})
				if !writeStatus {
					log.Fatalf("Error write file=%v", frameFileName)
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
				os.Remove(frameFileName)
				os.Remove(paletteFileName)
			}

			scaledVideoFrame.Close()
			paletteID := uuid.NewV4().String()
			for _, clr := range colors {
				result := &Result{
					SourceURL:             videoFilePath,
					SourceSerial:          "",
					SourceDurationSeconds: videoDuration,
					SourceFPS:             videoFps,
					SampleID:              periodID,
					SampleNumber:          period + 1,
					SampleDuration:        segmentDurationSeconds,
					PaletteCounts:         paletteSize,
					PaletteID:             paletteID,
				}
				result.R, result.G, result.B, result.A = clr.RGBA()
				result.Weight = p.Weight(clr)
				result.Normalize16BitRGB()
				results = append(results, result)
			}

		} else {
			elapsed := time.Since(start)
			log.Printf("Segment: %d took %s", period+1, elapsed)
			start = time.Now()
			frameCount = 0
			period++
			err = gocsv.MarshalFile(results, f)
			if err != nil {
				log.Fatalf("%v", err)
			}
			results = []*Result{}
			periodID = uuid.NewV4().String()
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

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
