# video-color-palette-generator

Generates color palette of a given video. Using k-means for clustering the color of video frames.

The output of this tool is a csv with the following structure

```
source_serial,source_url,source_duration_seconds,source_fps,sample_id,sample_number,sample_duration,palette_id,palette_counts,r,g,b,a,r_norm,g_norm,b_norm,weight
```

Data types for each attributes can be seen in this following struct 

```go
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
```

## Args

```
Usage: video-color-palette-generator.exe --input-file INPUT-FILE --period-duration PERIOD-DURATION --palette-size PALETTE-SIZE [--max-iteration MAX-ITERATION] --csv-result CSV-RESULT <command> [<args>]

Options:
  --input-file INPUT-FILE, -i INPUT-FILE
                         input file path for video
  --period-duration PERIOD-DURATION, -d PERIOD-DURATION
                         period duration in seconds
  --palette-size PALETTE-SIZE, -k PALETTE-SIZE
                         palette size
  --max-iteration MAX-ITERATION
                         maximum iteration of k-means if not convergent [default: 300]
  --csv-result CSV-RESULT, -o CSV-RESULT
                         csv result path
  --help, -h             display this help and exit

Commands:
  visualize
```

### Visualize command

```
Usage: video-color-palette-generator.exe visualize [--visualize-output-folder VISUALIZE-OUTPUT-FOLDER]

Options:
  --visualize-output-folder VISUALIZE-OUTPUT-FOLDER
                         visualization output folder. Contains frame and color palette

Global options:
  --input-file INPUT-FILE, -i INPUT-FILE
                         input file path for video
  --period-duration PERIOD-DURATION, -d PERIOD-DURATION
                         period duration in seconds
  --palette-size PALETTE-SIZE, -k PALETTE-SIZE
                         palette size
  --max-iteration MAX-ITERATION
                         maximum iteration of k-means if not convergent [default: 300]
  --csv-result CSV-RESULT, -o CSV-RESULT
                         csv result path
  --help, -h             display this help and exit
```

```shell
go run .\main.go .\t1.mp4 10 frames_output 6 300
```






