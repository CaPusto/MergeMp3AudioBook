// FILE: .\types.go
package main

import (
	"context"
	"os"
	"sync"
)

type Chapter struct {
	Title      string
	StartMs    int64
	EndMs      int64
	FileSizeMB float64
}

type ConverterApp struct {
	mu               sync.Mutex
	langMu           sync.RWMutex
	ffmpegPath       string
	ffprobePath      string
	t                map[string]string
	conversionCtx    context.Context
	cancelConversion context.CancelFunc
	conversionDone   chan struct{}
	logFile          *os.File
	uiLogLines       []string
	selectedBitrate  string
}
