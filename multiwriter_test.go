package cmw_test

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/lrweck/cmw"
)

var errBrokenWriter = errors.New("broken writer")

type brokerWriter struct{}

func (bw *brokerWriter) Write(p []byte) (int, error) {
	return len(p), errBrokenWriter
}

func TestMultiWriter(t *testing.T) {

	type test struct {
		label  string
		testFn func(t *testing.T)
	}

	tests := []test{
		{
			label: "TestMultiWriterSuccess",
			testFn: func(t *testing.T) {

				str := "This is a sample string"
				reader := strings.NewReader(str)

				var b1, b2, b3, b4, b5, b6 bytes.Buffer

				writer := cmw.ConcurrentMultiWriter(&b1, &b2, &b3, &b4, &b5, &b6)

				n, err := io.Copy(writer, reader)

				if err != nil {
					t.Errorf("Error while copying: %v", err)
					t.FailNow()
				}

				if n != int64(len(str)) {
					t.Errorf("Expected %d bytes, got %d", len(str), n)
					t.FailNow()
				}
			},
		},
		{
			label: "TestMultiWriterFailed",
			testFn: func(t *testing.T) {

				str := "This is a sample string"
				reader := strings.NewReader(str)

				var b1, b2, b3, b4, b5, b6 bytes.Buffer
				var broken brokerWriter

				writer := cmw.ConcurrentMultiWriter(&b1, &b2, &b3, &b4, &b5, &b6, &broken)

				n, err := io.Copy(writer, reader)

				if err == nil {
					t.Errorf("Expected error '%v', got nil", errBrokenWriter)
					t.FailNow()
				}

				if n != int64(0) {
					t.Errorf("Expected %d bytes, got %d", n, len(str))
					t.FailNow()
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.label, test.testFn)
	}

}

type slowWriter struct {
	delay time.Duration
}

func (sw *slowWriter) Write(p []byte) (int, error) {
	time.Sleep(sw.delay)
	return len(p), nil
}

func (sw *slowWriter) WriteStringWriteString(s string) (n int, err error) {
	time.Sleep(sw.delay)
	return len(s), nil
}

func BenchmarkMultiWriter(b *testing.B) {

	const testSize = 80000

	reader := strings.NewReader(strings.Repeat("a", testSize))

	writers := make([]io.Writer, 0, testSize)
	for i := 0; i < testSize; i++ {
		writers = append(writers, &slowWriter{delay: time.Millisecond * time.Duration(rand.Intn(10))})
	}

	b.Run("cmw.ConcurrentMultiWriter", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			writer := cmw.ConcurrentMultiWriter(writers...)
			_, _ = io.Copy(writer, reader)
		}
	})

	writers = make([]io.Writer, 0, testSize)
	for i := 0; i < testSize; i++ {
		writers = append(writers, &slowWriter{delay: time.Millisecond * time.Duration(rand.Intn(10))})
	}

	b.Run("io.MultiWriter", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			writer := io.MultiWriter(writers...)
			_, _ = io.Copy(writer, reader)
		}
	})

}
