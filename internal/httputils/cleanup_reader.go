package httputils

import "io"

// cleanupReader wraps io.Reader to detect when the read is complete and the file is downloaded
type cleanupReader struct {
	io.ReadSeeker
	cleanup func()
	done    bool
}

func (r *cleanupReader) Read(p []byte) (n int, err error) {
	n, err = r.ReadSeeker.Read(p)
	if err == io.EOF && !r.done {
		r.done = true
		r.cleanup()
	}
	return n, err
}

func (r *cleanupReader) Seek(offset int64, whence int) (int64, error) {
	// If we're seeking to the end and haven't called cleanup yet, do it now
	if whence == io.SeekEnd && offset == 0 && !r.done {
		r.done = true
		r.cleanup()
	}
	return r.ReadSeeker.Seek(offset, whence)
}

// NewCleanupReader creates a new cleanupReader
func NewCleanupReader(r io.ReadSeeker, cleanup func()) io.ReadSeeker {
	return &cleanupReader{
		ReadSeeker: r,
		cleanup:    cleanup,
	}
}
