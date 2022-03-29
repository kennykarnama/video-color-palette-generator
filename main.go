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

	"gitlab.com/ruangguru/kennykarnama/video-color-palette-generator/ffprobe"

	"github.com/gocarina/gocsv"

	"time"

	"github.com/alexflint/go-arg"
	gim "github.com/ozankasikci/go-image-merge"
	"github.com/satori/go.uuid"
	"gocv.io/x/gocv"
	ct "github.com/kennykarnama/color-thief"
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
	FunctionType   int `arg:"required,--function-type" help:"function type. 0 --> quant_wu, 1 --> WSM_WU"`
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
	videoDurationMs := prober.GetDuration().Milliseconds()
	frameCountsPerSegment := math.Floor(videoFps * segmentDurationSeconds)

	log.Printf("FPS=%v", videoFps)
	log.Printf("FrameCountsPerSegment=%v", frameCountsPerSegment)
	log.Printf("durationSeconds=%v", videoDuration)
	log.Printf("durationMs=%v", videoDurationMs)

	videoFrame := gocv.NewMat()
	defer videoFrame.Close()

	frameCount := float64(0)
	period := 0

	f, err := os.OpenFile(resultFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer f.Close()

	defer timeTrack(time.Now(), "video-color-palette-extraction")

	start := time.Now()

	var desiredIdx float64
	//totalFrames := vc.Get(gocv.VideoCaptureFrameCount)

	for desiredIdx = float64(0); desiredIdx <= float64(videoDurationMs); desiredIdx += (segmentDurationSeconds * 1000) {

		vc.Set(gocv.VideoCapturePosMsec, desiredIdx)
		
		var tmpImage image.Image
		var err error
		var imageExist bool

		start = time.Now()
		
		for {
			if ok := vc.Read(&videoFrame); !ok {
				log.Printf("Video frame closed")
				break
			}
			if videoFrame.Empty() {
				log.Printf("Video frame empty")
				continue
			}

			imageExist = true

			// scale frame
			scaledVideoFrame := gocv.NewMat()
			gocv.Resize(videoFrame, &scaledVideoFrame, image.Point{X: 0, Y: 0}, 0.1, 0.1, gocv.InterpolationCubic)
			
			// generate palette
			tmpImage, err = scaledVideoFrame.ToImage()
			if err != nil {
				scaledVideoFrame.Close()
				log.Fatalf("Error convert Mat to image err=%v", err)
			}
			scaledVideoFrame.Close()
			break
		}


		frameCount++

		if imageExist {

			colors, err := ct.GetPalette(tmpImage, int(paletteSize), args.FunctionType)
			if err != nil {
				log.Fatalf("Palettor err=%v", err)
			}

			if args.VisualizeCmd != nil {
				frameFileName := filepath.Join(outputFolder, fmt.Sprintf("frame_%v__segment_%v.png", frameCount, period+1))
				log.Printf("writing file=%v", frameFileName)
				writeStatus := gocv.IMWriteWithParams(frameFileName, videoFrame, []int{gocv.IMWritePngStrategy})
				if !writeStatus {
					log.Fatalf("Error write file=%v", frameFileName)
				}
				paletteFileName := filepath.Join(outputFolder, fmt.Sprintf("palette_%v__segment_%v.png", frameCount, period+1))
				log.Printf("writing palette file=%v", paletteFileName)

				_, err = createPalette(paletteFileName, colors)
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

			paletteID := uuid.NewV4().String()
			period++
			periodID := uuid.NewV4().String()
			var results []*Result
			for _, clr := range colors {
				result := &Result{
					SourceURL:             videoFilePath,
					SourceSerial:          "",
					SourceDurationSeconds: videoDuration,
					SourceFPS:             videoFps,
					SampleID:              periodID,
					SampleNumber:          period,
					SampleDuration:        segmentDurationSeconds,
					PaletteCounts:         len(colors),
					PaletteID:             paletteID,
				}
				result.R, result.G, result.B, result.A = clr.RGBA()
				result.Normalize16BitRGB()
				results = append(results, result)
			}
			if period == 1 {
				// write csv
				err = gocsv.MarshalFile(results, f)
				if err != nil {
					log.Fatalf("%v", err)
				}
			}else {
				// write csv
				err = gocsv.MarshalWithoutHeaders(results, f)
				if err != nil {
					log.Fatalf("%v", err)
				}
			}
			elapsed := time.Since(start)
			log.Printf("Segment: %d k=%v took=%s", period, len(colors), elapsed)

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
