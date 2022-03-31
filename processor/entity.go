package processor


type Parameter  struct {
	InputFile      string         `arg:"--input-file,-i" help:"input file path for video"`
	InputSerial    string `arg:"--input-serial" help:"input serial for video"`
	PeriodDuration float64        `arg:"--period-duration,-d" help:"period duration in seconds"`
	PaletteSize    int            `arg:"--palette-size,-k" help:"palette size"`
	FunctionType   int            `arg:"--function-type" help:"function type. 0 --> quant_wu, 1 --> WSM_WU"`
	CsvResult      string         `arg:"--csv-result,-o" help:"csv result path"`
	VisualizeCmd   *VisualizeArgs `arg:"subcommand:visualize"`
}


type VisualizeArgs struct {
	OutputFolder string `arg:"--visualize-output-folder" help:"visualization output folder. Contains frame and color palette"`
}

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

