package term

import (
	"context"
	"os"
	"sync"
)

const readBufSize = 4096

// Reader reads raw bytes from a terminal file descriptor in a background goroutine.
// The kernel's tty layer delivers escape sequences atomically in a single read,
// so no coalescing is needed.
type Reader struct {
	in   *os.File
	ch   chan []byte
	stop chan struct{}
	once sync.Once
}

// NewReader creates a Reader for the given input file.
func NewReader(in *os.File) *Reader {
	return &Reader{
		in:   in,
		ch:   make(chan []byte, 4),
		stop: make(chan struct{}),
	}
}

// Start begins the background read loop. It stops when ctx is cancelled
// or Stop is called.
func (r *Reader) Start(ctx context.Context) {
	go r.loop(ctx)
}

// Events returns the channel of raw byte chunks.
func (r *Reader) Events() <-chan []byte {
	return r.ch
}

// Stop signals the read loop to exit.
func (r *Reader) Stop() {
	r.once.Do(func() { close(r.stop) })
}

func (r *Reader) loop(ctx context.Context) {
	buf := make([]byte, readBufSize)
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stop:
			return
		default:
		}

		n, err := r.in.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])

			select {
			case r.ch <- chunk:
			case <-ctx.Done():
				return
			case <-r.stop:
				return
			}
		}
		if err != nil {
			return
		}
	}
}
