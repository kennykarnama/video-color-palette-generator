package lambdaapi

type ColorPaletteGenerationRequest struct {
	SourceURL string `json:"sourceURL"`
	SourceSerial string `json:"sourceSerial"`
	DestinationURI string `json:"destinationURI"`
}
