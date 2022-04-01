package processor


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

	"github.com/kennykarnama/video-color-palette-generator/ffprobe"

	"github.com/gocarina/gocsv"

	"time"

	gim "github.com/ozankasikci/go-image-merge"
	"github.com/satori/go.uuid"
	"gocv.io/x/gocv"
	ct "github.com/kennykarnama/color-thief"
)


func Run(args Parameter) error {

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
		return fmt.Errorf("action=run.video_capture_file path=%v err=%v", videoFilePath, err)
	}
	defer vc.Close()

	prober, err := ffprobe.NewFfprobe(videoFilePath)
	if err != nil {
		return fmt.Errorf("action=run.ffprobe path=%v err=%v", videoFilePath, err)
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
		return fmt.Errorf("action=run.open_result_file path=%v result_file=%v err=%v", videoFilePath, resultFilePath, err)
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
				return fmt.Errorf("action=run.scaledVideoFrameToImage err=%v", err)
			}
			scaledVideoFrame.Close()
			break
		}


		frameCount++

		if imageExist {

			colors, err := ct.GetPalette(tmpImage, int(paletteSize), args.FunctionType)
			if err != nil {
				return fmt.Errorf("action=run.GetPalette err=%v", err)
			}

			if args.VisualizeCmd != nil {
				frameFileName := filepath.Join(outputFolder, fmt.Sprintf("frame_%v__segment_%v.png", frameCount, period+1))
				log.Printf("writing file=%v", frameFileName)
				writeStatus := gocv.IMWriteWithParams(frameFileName, videoFrame, []int{gocv.IMWritePngStrategy})
				if !writeStatus {
					return fmt.Errorf("action=run.WriteVideoFrame target=%v err=%v", frameFileName, err)
				}
				paletteFileName := filepath.Join(outputFolder, fmt.Sprintf("palette_%v__segment_%v.png", frameCount, period+1))
				log.Printf("writing palette file=%v", paletteFileName)

				_, err = createPalette(paletteFileName, colors)
				if err != nil {
					return fmt.Errorf("action=run.createPaletteFile target=%v err=%v", paletteFileName, err)
				}

				visualizeFileName := filepath.Join(outputFolder, fmt.Sprintf("visualize_%v__segment_%v.png", frameCount, period+1))
				// merge frame and palette to allow better visualization
				err = visualize(frameFileName, paletteFileName, visualizeFileName)
				if err != nil {
					return fmt.Errorf("action=run.createVisualizeFile target=%v err=%v", visualizeFileName, err)
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
					SourceSerial:          args.InputSerial,
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
					return fmt.Errorf("action=run.csv_marshal err=%v", err)
				}
			}else {
				// write csv
				err = gocsv.MarshalWithoutHeaders(results, f)
				if err != nil {
					return fmt.Errorf("action=run.csv_marshalWithoutHeaders err=%v", err)
				}
			}
			elapsed := time.Since(start)
			log.Printf("Segment: %d k=%v took=%s", period, len(colors), elapsed)

		}
	}
	return nil
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
