# video-color-palette-generator

Generates color palette of a given video. Using k-means for clustering the color of video frames.

This tool has two modes:

- lambdaHandler
- script

# Modes

## Lambbda

As lambda, this tool introduces two phases:

- parse lambda event request
- run `processor (script)`

### Lambda Deployment

For deployment to lambda, you'll need to build docker image

Follow steps explained here: https://docs.aws.amazon.com/AmazonECR/latest/userguide/docker-push-ecr-image.html

Then you need to create lambda function from docker image explained here: https://docs.aws.amazon.com/lambda/latest/dg/images-create.html

## Script

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

### Args

For args, please run `./video-color-palette-generator script --help`

# Thanks

Big thanks for open source project here:

- https://github.com/hybridgroup/gocv
- https://github.com/zluo01/color-thief





