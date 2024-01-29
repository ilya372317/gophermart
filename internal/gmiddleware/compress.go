package gmiddleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ilya372317/gophermart/internal/logger"
)

const (
	contentEncodingHeader = "Content-Encoding"
	gzipEncoding          = "gzip"
)

const (
	LastPositiveStatusCode       = 300
	failedCompressDataErrPattern = "failed compress data: %w"
)

type Writer struct {
	w          http.ResponseWriter
	gzipWriter *gzip.Writer
}

func newWriter(writer http.ResponseWriter) *Writer {
	gzipWriter := gzip.NewWriter(writer)

	return &Writer{
		w:          writer,
		gzipWriter: gzipWriter,
	}
}

func (w *Writer) Header() http.Header {
	return w.w.Header()
}

func (w *Writer) Write(bytes []byte) (int, error) {
	size, err := w.gzipWriter.Write(bytes)
	if err != nil {
		err = fmt.Errorf(failedCompressDataErrPattern, err)
	}
	return size, err
}

func (w *Writer) WriteHeader(statusCode int) {
	if statusCode < LastPositiveStatusCode {
		w.w.Header().Set("Content-Encoding", "gzip")
	}

	w.w.WriteHeader(statusCode)
}

func (w *Writer) Close() error {
	err := w.gzipWriter.Close()
	if err != nil {
		err = fmt.Errorf("failed close gzip response writer: %w", err)
	}
	return err
}

type Reader struct {
	r          io.ReadCloser
	gzipReader *gzip.Reader
}

func (r *Reader) Read(p []byte) (n int, err error) {
	return r.gzipReader.Read(p) //nolint //error may be good case
}

func (r *Reader) Close() error {
	err := r.r.Close()
	if err != nil {
		return fmt.Errorf("failed close response reader: %w", err)
	}

	err = r.gzipReader.Close()
	if err != nil {
		err = fmt.Errorf("failed close gzip reader: %w", err)
	}

	return err
}

func newReader(reader io.ReadCloser) (*Reader, error) {
	gReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed create new gzip reader: %w", err)
	}

	return &Reader{
		r:          reader,
		gzipReader: gReader,
	}, nil
}

func Compressed() func(handler http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			acceptGzip := strings.Contains(acceptEncoding, gzipEncoding)
			if acceptGzip {
				cw := newWriter(w)
				ow = cw
				w.Header().Set(contentEncodingHeader, gzipEncoding)
				defer func() {
					_ = cw.Close()
				}()
			}

			contentEncoding := r.Header.Get(contentEncodingHeader)
			contentCompressed := strings.Contains(contentEncoding, gzipEncoding)
			if contentCompressed {
				cr, err := newReader(r.Body)
				if err != nil {
					http.Error(w, "failed create gzip compressor", http.StatusInternalServerError)
					logger.Log.Warnf("something went wrong with gzip compressor: %v", err)
					return
				}
				r.Body = cr
				defer func() {
					_ = cr.Close()
				}()
			}

			h.ServeHTTP(ow, r)
		})
	}
}
