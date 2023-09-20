package test

import (
	"bufio"
	"fmt"
	"testing"
)

type Writer int

func (w *Writer) Write(p []byte) (n int, err error) {
	fmt.Println(len(p))
	return len(p), nil
}

func TestBufferSize(t *testing.T) {
	fmt.Println("Unbuffered I/O")
	w := new(Writer)
	w.Write([]byte{'a'})
	w.Write([]byte{'b'})
	w.Write([]byte{'c'})
	w.Write([]byte{'d'})

	fmt.Println("Buffered I/O")
	bw := bufio.NewWriterSize(w, 3)
	bw.Write([]byte{'a'})
	bw.Write([]byte{'b'})
	bw.Write([]byte{'c'})
	bw.Write([]byte{'d'})
	err := bw.Flush()
	if err != nil {
		t.Error(err)
	}
}

func TestLargeWrite(t *testing.T) {
	w := new(Writer)
	bw := bufio.NewWriterSize(w, 3)
	// 如果 Writer 检测到 Write 方法被调用时传入的数据长度大于缓存的长度(示例中是三个字节)。其将直接调用 writer(目的对象)的 Write 方法
	bw.Write([]byte("abcd"))
}

type Writer1 int

func (w *Writer1) Write(p []byte) (n int, err error) {
	fmt.Printf("writer#1: %q\n", p)
	return len(p), nil
}

type Writer2 int

func (w *Writer2) Write(p []byte) (n int, err error) {
	fmt.Printf("writer#2: %q\n", p)
	return len(p), nil
}

func TestWriteReset(t *testing.T) {
	w1 := new(Writer1)
	bw := bufio.NewWriterSize(w1, 2)
	bw.Write([]byte("ab"))
	bw.Write([]byte("cd"))
	// bug: bw.Flush()

	w2 := new(Writer2)
	// 由于 Reset 只是简单的丢弃未被处理的数据，所以已经被写入的数据 cd 丢失了
	bw.Reset(w2)
	bw.Write([]byte("ef"))
	bw.Flush()
}

// 检测缓存中还剩余多少空间
func TestWriteAvailable(t *testing.T) {
	w := new(Writer)
	bw := bufio.NewWriterSize(w, 2)
	fmt.Println(bw.Available())
	bw.Write([]byte{'a'})
	fmt.Println(bw.Available())
	bw.Write([]byte{'b'})
	fmt.Println(bw.Available())
	bw.Write([]byte{'c'})
	fmt.Println(bw.Available())
}
