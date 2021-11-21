package internal

import _ "embed"

var (
	//go:embed success.json
	ExampleSuccess []byte

	//go:embed failure.json
	ExampleFailure []byte
)
