package rwutil

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"sync"

	"braces.dev/errtrace"
)

func Scan(r io.Reader, a ...*byte) error {
	size := len(a)
	buf := make([]byte, size)

	_, err := io.ReadFull(r, buf)
	if err != nil {
		return errtrace.Wrap(err)
	}

	for i := range a {
		*a[i] = buf[i]
	}

	return nil
}

func ScanBuf(r io.Reader, size int) ([]byte, error) {
	buf := make([]byte, size)

	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	return buf, nil
}

func WriteBytesFlush(w *bufio.Writer, b ...[]byte) error {
	for _, arr := range b {
		_, err := w.Write(arr)
		if err != nil {
			return errtrace.Wrap(err)
		}
	}

	return errtrace.Wrap(w.Flush())
}

func WriteStringFlush(w *bufio.Writer, s string) error {
	_, err := w.WriteString(s)
	if err != nil {
		return errtrace.Wrap(err)
	}

	return errtrace.Wrap(w.Flush())
}

func WriteResponseFlush(w *bufio.Writer, r http.Response) error {
	err := r.Write(w)
	if err != nil {
		return errtrace.Wrap(err)
	}

	return errtrace.Wrap(w.Flush())
}

func TunnelConns(a, b net.Conn) {
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer a.Close()
		defer wg.Done()
		io.Copy(a, b)
	}()

	go func() {
		defer b.Close()
		defer wg.Done()
		io.Copy(b, a)
	}()

	wg.Wait()
}
