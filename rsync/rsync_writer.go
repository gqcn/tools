package main

import (
	"bytes"
	"sync"

	"github.com/gogf/gf/v2/text/gstr"
)

type Writer struct {
	mu     sync.RWMutex
	buffer *bytes.Buffer
}

func NewWriter() *Writer {
	return &Writer{
		mu:     sync.RWMutex{},
		buffer: bytes.NewBuffer(nil),
	}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Write(p)
}

func (w *Writer) Reset() {
	w.buffer.Reset()
}

func (w *Writer) String() string {
	return gstr.Trim(w.buffer.String())
}
