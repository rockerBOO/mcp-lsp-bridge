package lsp

import (
	"io"
	"sync"
)

// stdioReadWriteCloser implements io.ReadWriteCloser
type stdioReadWriteCloser struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	
	mu     sync.Mutex
	closed bool
}

func (rwc *stdioReadWriteCloser) Read(p []byte) (n int, err error) {
	rwc.mu.Lock()
	defer rwc.mu.Unlock()
	
	if rwc.closed {
		return 0, io.EOF
	}
	
	return rwc.stdout.Read(p)
}

func (rwc *stdioReadWriteCloser) Write(p []byte) (n int, err error) {
	rwc.mu.Lock()
	defer rwc.mu.Unlock()
	
	if rwc.closed {
		return 0, io.ErrClosedPipe
	}
	
	return rwc.stdin.Write(p)
}

func (rwc *stdioReadWriteCloser) Close() error {
	rwc.mu.Lock()
	defer rwc.mu.Unlock()
	
	if rwc.closed {
		return nil // Already closed, not an error
	}
	
	rwc.closed = true
	
	var errors []error
	
	if err := rwc.stdin.Close(); err != nil {
		errors = append(errors, err)
	}
	
	if err := rwc.stdout.Close(); err != nil {
		errors = append(errors, err)
	}
	
	// Return the first error if any
	if len(errors) > 0 {
		return errors[0]
	}
	
	return nil
}
