package main

import (
	"log"

	"github.com/alexflint/go-arg"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/kennykarnama/video-color-palette-generator/processor"
	"github.com/kennykarnama/video-color-palette-generator/lambdaapi"
)

var args struct {
	ScriptCmd *processor.Parameter `arg:"subcommand:script"`
	LambdaCmd *LambdaArgs `arg:"subcommand:lambda"`
}

type LambdaArgs struct{}


type VisualizeArgs struct {
	OutputFolder string `arg:"--visualize-output-folder" help:"visualization output folder. Contains frame and color palette"`
}

func main() {
	// parse args
	arg.MustParse(&args)

	if args.LambdaCmd != nil {
		log.Printf("Running as lambda")
		lambda.Start(lambdaapi.ColorPaletteHandler)
	}else {
		log.Printf("Running as script")
		err := processor.Run(*args.ScriptCmd)
		if err != nil {
			log.Fatalf("err=%v", err)
		}
	}
}

