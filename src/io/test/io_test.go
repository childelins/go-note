package test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestReader(t *testing.T) {
	reader := strings.NewReader("Clear is better than clever")
	p := make([]byte, 4)

	for {
		n, err := reader.Read(p)
		if err != nil {
			if err == io.EOF {
				fmt.Println("EOF:", n)
				break
			}

			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(n, string(p[:n]))
	}
}

type alphaReader struct {
	r io.Reader
}

func newAlphaReader(r io.Reader) *alphaReader {
	return &alphaReader{r: r}
}

func (a *alphaReader) Read(p []byte) (int, error) {
	n, err := a.r.Read(p)
	if err != nil {
		return n, err
	}

	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		if c := alpha(p[i]); c != 0 {
			buf[i] = c
		}
	}

	copy(p, buf)
	return n, nil
}

func alpha(r byte) byte {
	if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
		return r
	}
	return 0
}

func TestAlphaReader(t *testing.T) {
	reader := newAlphaReader(strings.NewReader("Hello! It's 9am, where is the sun?"))
	p := make([]byte, 4)

	for {
		n, err := reader.Read(p)
		if err == io.EOF {
			break
		}

		fmt.Print(string(p[:n]))
	}
	fmt.Println()
}

func TestAlphaReaderFromFile(t *testing.T) {
	file, err := os.Open("./alpha_reader.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	reader := newAlphaReader(file)
	p := make([]byte, 4)

	for {
		n, err := reader.Read(p)
		if err == io.EOF {
			break
		}

		fmt.Print(string(p[:n]))
	}
	fmt.Println()
}

func TestWriter(t *testing.T) {
	proverbs := []string{
		"Channels orchestrate mutexes serialize",
		"Cgo is not Go",
		"Errors are values",
		"Don't panic",
	}

	var writer bytes.Buffer

	for _, p := range proverbs {
		n, err := writer.Write([]byte(p))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if n != len(p) {
			fmt.Println("failed to write data")
			os.Exit(1)
		}
	}

	fmt.Println(writer.String())
}

type chanWriter struct {
	ch chan byte
}

func newChanWriter() *chanWriter {
	return &chanWriter{ch: make(chan byte, 1024)}
}

func (w *chanWriter) Write(p []byte) (int, error) {
	n := 0
	for _, b := range p {
		w.ch <- b
		n++
	}
	return n, nil
}

func (w *chanWriter) Chan() <-chan byte {
	return w.ch
}

func (w *chanWriter) Close() error {
	close(w.ch)
	return nil
}

func TestChanWriter(t *testing.T) {
	writer := newChanWriter()

	go func() {
		defer writer.Close()
		writer.Write([]byte("Stream "))
		writer.Write([]byte("me!"))
	}()

	for c := range writer.Chan() {
		fmt.Printf("%c", c)
	}

	fmt.Println()
}

func TestWriteToFile(t *testing.T) {
	proverbs := []string{
		"Channels orchestrate mutexes serialize\n",
		"Cgo is not Go\n",
		"Errors are values\n",
		"Don't panic\n",
	}

	file, err := os.Create("./proverbs.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	for _, p := range proverbs {
		n, err := file.Write([]byte(p))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if n != len(p) {
			fmt.Println("failed to write data")
			os.Exit(1)
		}
	}

	fmt.Println("file write done")
}

func TestWriteToStd(t *testing.T) {
	proverbs := []string{
		"Channels orchestrate mutexes serialize\n",
		"Cgo is not Go\n",
		"Errors are values\n",
		"Don't panic\n",
	}

	for _, p := range proverbs {
		n, err := os.Stdout.Write([]byte(p))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if n != len(p) {
			fmt.Println("failed to write data")
			os.Exit(1)
		}
	}
}

func TestWriteToFileWithCopy(t *testing.T) {
	proverbs := new(bytes.Buffer)
	proverbs.WriteString("Channels orchestrate mutexes serialize\n")
	proverbs.WriteString("Cgo is not Go\n")
	proverbs.WriteString("Errors are values\n")
	proverbs.WriteString("Don't panic\n")

	file, err := os.Create("./proverbs2.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	if _, err := io.Copy(file, proverbs); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("file write done")
}
