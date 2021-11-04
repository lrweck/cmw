package cmw

import (
	"io"

	"golang.org/x/sync/errgroup"
)

type concurrentMultiWriter struct {
	writers []io.Writer
}

func (t *concurrentMultiWriter) Write(p []byte) (n int, err error) {

	g := errgroup.Group{}

	for _, w := range t.writers {
		w := w
		g.Go(func() error {
			v := 0
			v, err = w.Write(p)
			if err != nil {
				return err
			}
			if v != len(p) {
				return io.ErrShortWrite
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return 0, err
	}

	return len(p), nil
}

var _ io.StringWriter = (*concurrentMultiWriter)(nil)

func (t *concurrentMultiWriter) WriteString(s string) (n int, err error) {

	g := errgroup.Group{}

	for _, w := range t.writers {
		w := w
		g.Go(func() error {
			v := 0
			if sw, ok := w.(io.StringWriter); ok {
				v, err = sw.WriteString(s)
			} else {
				v, err = w.Write([]byte(s))
			}
			if err != nil {
				return err
			}
			if v != len(s) {
				return io.ErrShortWrite
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return 0, err
	}

	return len(s), nil
}

// ConcurrentMultiWriter creates a writer that duplicates its writes to all the
// provided writers, similar to the Unix tee(1) command.
//
// Each write is written to each listed writer, concurrently.
// If a listed writer returns an error, that overall write operation
// stops and returns the error; it does not continue down the list.
func ConcurrentMultiWriter(writers ...io.Writer) io.Writer {
	allWriters := make([]io.Writer, 0, len(writers))
	for _, w := range writers {
		if mw, ok := w.(*concurrentMultiWriter); ok {
			allWriters = append(allWriters, mw.writers...)
		} else {
			allWriters = append(allWriters, w)
		}
	}
	return &concurrentMultiWriter{allWriters}
}
