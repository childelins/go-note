// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package io provides basic interfaces to I/O primitives.
// Its primary job is to wrap existing implementations of such primitives,
// such as those in package os, into shared public interfaces that
// abstract the functionality, plus some other related primitives.
//
// Because these interfaces and primitives wrap lower-level operations with
// various implementations, unless otherwise informed clients should not
// assume they are safe for parallel execution.

// io包提供了与I/O原语相关的基本接口。它的主要工作是将现有的原语实现（例如包os中的实现）封装成共享的公共接口，该接口抽象了功能，以及一些其他相关的原语。

// 因为这些接口和原语使用各种实现来封装低级操作，除非另有通知，客户端不应假设它们适用于并行执行。

// 参考：https://segmentfault.com/a/1190000015591319
package io

import (
	"errors"
	"sync"
)

// Seek whence values.
const (
	SeekStart   = 0 // seek relative to the origin of the file
	SeekCurrent = 1 // seek relative to the current offset
	SeekEnd     = 2 // seek relative to the end
)

// ErrShortWrite means that a write accepted fewer bytes than requested
// but failed to return an explicit error.
var ErrShortWrite = errors.New("short write")

// errInvalidWrite means that a write returned an impossible count.
var errInvalidWrite = errors.New("invalid write result")

// ErrShortBuffer means that a read required a longer buffer than was provided.
var ErrShortBuffer = errors.New("short buffer")

// EOF is the error returned by Read when no more input is available.
// (Read must return EOF itself, not an error wrapping EOF,
// because callers will test for EOF using ==.)
// Functions should return EOF only to signal a graceful end of input.
// If the EOF occurs unexpectedly in a structured data stream,
// the appropriate error is either ErrUnexpectedEOF or some other error
// giving more detail.
var EOF = errors.New("EOF")

// ErrUnexpectedEOF means that EOF was encountered in the
// middle of reading a fixed-size block or data structure.
var ErrUnexpectedEOF = errors.New("unexpected EOF")

// ErrNoProgress is returned by some clients of a Reader when
// many calls to Read have failed to return any data or error,
// usually the sign of a broken Reader implementation.
var ErrNoProgress = errors.New("multiple Read calls return no data or error")

// Reader is the interface that wraps the basic Read method.
//
// Read reads up to len(p) bytes into p. It returns the number of bytes
// read (0 <= n <= len(p)) and any error encountered. Even if Read
// returns n < len(p), it may use all of p as scratch space during the call.
// If some data is available but not len(p) bytes, Read conventionally
// returns what is available instead of waiting for more.
//
// When Read encounters an error or end-of-file condition after
// successfully reading n > 0 bytes, it returns the number of
// bytes read. It may return the (non-nil) error from the same call
// or return the error (and n == 0) from a subsequent call.
// An instance of this general case is that a Reader returning
// a non-zero number of bytes at the end of the input stream may
// return either err == EOF or err == nil. The next Read should
// return 0, EOF.
//
// Callers should always process the n > 0 bytes returned before
// considering the error err. Doing so correctly handles I/O errors
// that happen after reading some bytes and also both of the
// allowed EOF behaviors.
//
// Implementations of Read are discouraged from returning a
// zero byte count with a nil error, except when len(p) == 0.
// Callers should treat a return of 0 and nil as indicating that
// nothing happened; in particular it does not indicate EOF.
//
// Implementations must not retain p.

// Reader是一个接口，它封装了基本的Read方法。

// Read方法将最多len(p)个字节读入p中。它返回读取的字节数（0 <= n <= len(p)）和遇到的任何错误。即使Read返回的n < len(p)，在调用期间它也可以使用p的所有字节作为临时空间。如果有一些数据可用但不足len(p)字节，Read通常会返回可用的数据而不是等待更多数据。

// 当Read在成功读取n > 0个字节后遇到错误或文件结束条件时，它返回读取的字节数。它可以从同一次调用返回（非nil）错误，也可以从后续调用返回错误（和n == 0）。一个常见的情况是，一个Reader在输入流的末尾返回非零字节数，可能会返回err EOF或err == nil。下一次Read应该返回0，EOF。

// 调用者应始终在考虑错误err之前处理返回的n > 0字节。这样做可以正确处理在读取一些字节之后发生的I/O错误以及两种允许的EOF行为。

// 不鼓励Read的实现在没有错误的情况下返回零字节计数，除非len(p) == 0。调用者应将返回的0和nil视为表示没有发生任何事情；特别是它不表示EOF。
type Reader interface {
	Read(p []byte) (n int, err error)
}

// Writer is the interface that wraps the basic Write method.
//
// Write writes len(p) bytes from p to the underlying data stream.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
// Write must not modify the slice data, even temporarily.
//
// Implementations must not retain p.

// Writer 是一个接口，它封装了基本的 Write 方法。

// Write 方法从 p 中将 len(p) 个字节写入到底层的数据流中。它返回从 p 中写入的字节数 n（0 <= n <= len(p)）以及导致写入提前停止的任何错误。如果 Write 返回 n < len(p)，则必须返回非空的错误。Write 方法不能修改切片数据，即使是暂时的修改。
type Writer interface {
	Write(p []byte) (n int, err error)
}

// Closer is the interface that wraps the basic Close method.
//
// The behavior of Close after the first call is undefined.
// Specific implementations may document their own behavior.

// Closer 是一个接口，它封装了基本的 Close 方法。

// 在第一次调用 Close 后，其行为是未定义的。具体的实现可能会对其行为进行文档化。
type Closer interface {
	Close() error
}

// Seeker is the interface that wraps the basic Seek method.
//
// Seek sets the offset for the next Read or Write to offset,
// interpreted according to whence:
// SeekStart means relative to the start of the file,
// SeekCurrent means relative to the current offset, and
// SeekEnd means relative to the end
// (for example, offset = -2 specifies the penultimate byte of the file).
// Seek returns the new offset relative to the start of the
// file or an error, if any.
//
// Seeking to an offset before the start of the file is an error.
// Seeking to any positive offset may be allowed, but if the new offset exceeds
// the size of the underlying object the behavior of subsequent I/O operations
// is implementation-dependent.

// Seeker 是一个接口，它封装了基本的 Seek 方法。

// Seek 方法用于设置下一次读取或写入的偏移量(offset)，根据 whence 参数的不同进行解释：
// 	SeekStart 表示相对于文件的开头进行偏移
// 	SeekCurrent 表示相对于当前偏移进行偏移
// 	SeekEnd 表示相对于文件的结尾进行偏移
// 例如，offset = -2 表示文件中倒数第二个字节
// Seek 方法返回相对于文件开头的新的偏移量或者错误（如果有的话）。

// 尝试将偏移量设置为文件开始之前是一个错误。将偏移量设置为任何正数可能是允许的，但如果新的偏移量超过了底层对象的大小，则后续的I/O操作的行为取决于具体的实现。
type Seeker interface {
	Seek(offset int64, whence int) (int64, error)
}

// ReadWriter is the interface that groups the basic Read and Write methods.
type ReadWriter interface {
	Reader
	Writer
}

// ReadCloser is the interface that groups the basic Read and Close methods.
type ReadCloser interface {
	Reader
	Closer
}

// WriteCloser is the interface that groups the basic Write and Close methods.
type WriteCloser interface {
	Writer
	Closer
}

// ReadWriteCloser is the interface that groups the basic Read, Write and Close methods.
type ReadWriteCloser interface {
	Reader
	Writer
	Closer
}

// ReadSeeker is the interface that groups the basic Read and Seek methods.
type ReadSeeker interface {
	Reader
	Seeker
}

// ReadSeekCloser is the interface that groups the basic Read, Seek and Close
// methods.
type ReadSeekCloser interface {
	Reader
	Seeker
	Closer
}

// WriteSeeker is the interface that groups the basic Write and Seek methods.
type WriteSeeker interface {
	Writer
	Seeker
}

// ReadWriteSeeker is the interface that groups the basic Read, Write and Seek methods.
type ReadWriteSeeker interface {
	Reader
	Writer
	Seeker
}

// ReaderFrom is the interface that wraps the ReadFrom method.
//
// ReadFrom reads data from r until EOF or error.
// The return value n is the number of bytes read.
// Any error except EOF encountered during the read is also returned.
//
// The Copy function uses ReaderFrom if available.

// ReaderFrom是一个接口，它包装了ReadFrom方法。
// ReadFrom方法从r中读取数据直到EOF或出现错误。
// 返回值n是读取的字节数。
// 任何读取过程中遇到的错误(除了EOF)也会被返回。
// 如果可用，Copy函数会使用ReaderFrom
type ReaderFrom interface {
	ReadFrom(r Reader) (n int64, err error)
}

// WriterTo is the interface that wraps the WriteTo method.
//
// WriteTo writes data to w until there's no more data to write or
// when an error occurs. The return value n is the number of bytes
// written. Any error encountered during the write is also returned.
//
// The Copy function uses WriterTo if available.

// WriterTo是一个接口，它包装了WriteTo方法。
// WriteTo方法将数据写入w直到没有更多数据可写入或发生错误。返回值n是写入的字节数。任何写入过程中遇到的错误也会被返回。
// 如果可用，Copy函数会使用WriterTo。
type WriterTo interface {
	WriteTo(w Writer) (n int64, err error)
}

// ReaderAt is the interface that wraps the basic ReadAt method.
//
// ReadAt reads len(p) bytes into p starting at offset off in the
// underlying input source. It returns the number of bytes
// read (0 <= n <= len(p)) and any error encountered.
//
// When ReadAt returns n < len(p), it returns a non-nil error
// explaining why more bytes were not returned. In this respect,
// ReadAt is stricter than Read.
//
// Even if ReadAt returns n < len(p), it may use all of p as scratch
// space during the call. If some data is available but not len(p) bytes,
// ReadAt blocks until either all the data is available or an error occurs.
// In this respect ReadAt is different from Read.
//
// If the n = len(p) bytes returned by ReadAt are at the end of the
// input source, ReadAt may return either err == EOF or err == nil.
//
// If ReadAt is reading from an input source with a seek offset,
// ReadAt should not affect nor be affected by the underlying
// seek offset.
//
// Clients of ReadAt can execute parallel ReadAt calls on the
// same input source.
//
// Implementations must not retain p.

// ReaderAt是一个接口，它包装了基本的ReadAt方法。

// ReadAt方法从底层输入源的偏移量off开始，将len(p)个字节读入到p中。它返回读取的字节数n（0 <= n <= len(p))和遇到的任何错误。

// 当ReadAt返回的n < len(p)时，它会返回一个非空的错误，解释为什么没有返回更多的字节。在这方面，ReadAt比Read更严格。

// 即使ReadAt返回的n < len(p)，在调用期间它也可能使用p作为临时空间。如果有一些数据可用但不足len(p)字节，ReadAt会阻塞，直到所有数据可用或发生错误。在这方面，ReadAt与Read不同。

// 如果ReadAt从输入源的末尾读取了n = len(p)个字节，则ReadAt可能返回err == EOF或err == nil。

// 如果ReadAt从具有寻址偏移量的输入源中读取，则ReadAt不应影响底层的寻址偏移量，也不应受其影响。

// ReadAt的客户端可以在相同的输入源上并行执行ReadAt调用。
type ReaderAt interface {
	ReadAt(p []byte, off int64) (n int, err error)
}

// WriterAt is the interface that wraps the basic WriteAt method.
//
// WriteAt writes len(p) bytes from p to the underlying data stream
// at offset off. It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// WriteAt must return a non-nil error if it returns n < len(p).
//
// If WriteAt is writing to a destination with a seek offset,
// WriteAt should not affect nor be affected by the underlying
// seek offset.
//
// Clients of WriteAt can execute parallel WriteAt calls on the same
// destination if the ranges do not overlap.
//
// Implementations must not retain p.

// WriterAt是一个接口，它包装了基本的WriteAt方法。

// WriteAt方法将p中的len(p)个字节从偏移量off写入到底层数据流中。它返回从p中写入的字节数n（0 <= n <= len(p))以及导致写入提前停止的任何错误。如果WriteAt返回n < len(p)，则必须返回一个非空的错误。
// 如果WriteAt将数据写入具有寻址偏移量的目标位置，则WriteAt不应影响底层的寻址偏移量，也不应受其影响。
// 如果WriteAt的客户端对相同的目标位置进行并行的WriteAt调用，并且这些范围没有重叠，那么它们是可以执行的。
type WriterAt interface {
	WriteAt(p []byte, off int64) (n int, err error)
}

// ByteReader is the interface that wraps the ReadByte method.
//
// ReadByte reads and returns the next byte from the input or
// any error encountered. If ReadByte returns an error, no input
// byte was consumed, and the returned byte value is undefined.
//
// ReadByte provides an efficient interface for byte-at-time
// processing. A Reader that does not implement  ByteReader
// can be wrapped using bufio.NewReader to add this method.

// ByteReader是一个接口，它包装了ReadByte方法。
// ReadByte方法从输入中读取并返回下一个字节，以及遇到的任何错误。如果ReadByte返回一个错误，则没有消耗输入字节，并且返回的字节值是未定义的。
// ReadByte提供了一种逐字节处理的高效接口。如果一个Reader没有实现ByteReader，可以使用bufio.NewReader对其进行包装以添加这个方法。
type ByteReader interface {
	ReadByte() (byte, error)
}

// ByteScanner is the interface that adds the UnreadByte method to the
// basic ReadByte method.
//
// UnreadByte causes the next call to ReadByte to return the last byte read.
// If the last operation was not a successful call to ReadByte, UnreadByte may
// return an error, unread the last byte read (or the byte prior to the
// last-unread byte), or (in implementations that support the Seeker interface)
// seek to one byte before the current offset.

// ByteScanner是一个接口，它在基本的ReadByte方法上添加了UnreadByte方法。
// UnreadByte方法会导致下一次调用ReadByte返回上一次读取的字节。如果上一次操作不是对ReadByte的成功调用，UnreadByte可能会返回一个错误，未读取上一个字节（或上一个未读取字节之前的字节），或者（在支持Seeker接口的实现中）将偏移量定位到当前偏移量之前的一个字节位置。
type ByteScanner interface {
	ByteReader
	UnreadByte() error
}

// ByteWriter is the interface that wraps the WriteByte method.
// ByteWriter是一个接口，它包装了WriteByte方法。
type ByteWriter interface {
	WriteByte(c byte) error
}

// RuneReader is the interface that wraps the ReadRune method.
//
// ReadRune reads a single encoded Unicode character
// and returns the rune and its size in bytes. If no character is
// available, err will be set.

// RuneReader是一个接口，它包装了ReadRune方法。
// ReadRune方法读取一个编码的Unicode字符，并返回该字符以及它在字节中的大小。如果没有可用的字符，err会被设置。
type RuneReader interface {
	ReadRune() (r rune, size int, err error)
}

// RuneScanner is the interface that adds the UnreadRune method to the
// basic ReadRune method.
//
// UnreadRune causes the next call to ReadRune to return the last rune read.
// If the last operation was not a successful call to ReadRune, UnreadRune may
// return an error, unread the last rune read (or the rune prior to the
// last-unread rune), or (in implementations that support the Seeker interface)
// seek to the start of the rune before the current offset.

// RuneScanner是一个接口，它在基本的ReadRune方法上添加了UnreadRune方法。
// UnreadRune方法会导致下一次调用ReadRune返回上一次读取的符文（rune）。如果上一次操作不是对ReadRune的成功调用，UnreadRune可能会返回一个错误，未读取上一个符文（或上一个未读取符文之前的符文），或者（在支持Seeker接口的实现中）将偏移量定位到当前偏移量之前的一个符文位置。
type RuneScanner interface {
	RuneReader
	UnreadRune() error
}

// StringWriter is the interface that wraps the WriteString method.
// StringWriter是一个接口，它包装了WriteString方法。
type StringWriter interface {
	WriteString(s string) (n int, err error)
}

// WriteString writes the contents of the string s to w, which accepts a slice of bytes.
// If w implements StringWriter, its WriteString method is invoked directly.
// Otherwise, w.Write is called exactly once.

// WriteString方法将字符串s的内容写入到接受字节切片的w中。
// 如果w实现了StringWriter接口，它的WriteString方法会直接被调用。
// 否则，w.Write方法会被调用一次。
func WriteString(w Writer, s string) (n int, err error) {
	if sw, ok := w.(StringWriter); ok {
		return sw.WriteString(s)
	}
	return w.Write([]byte(s))
}

// ReadAtLeast reads from r into buf until it has read at least min bytes.
// It returns the number of bytes copied and an error if fewer bytes were read.
// The error is EOF only if no bytes were read.
// If an EOF happens after reading fewer than min bytes,
// ReadAtLeast returns ErrUnexpectedEOF.
// If min is greater than the length of buf, ReadAtLeast returns ErrShortBuffer.
// On return, n >= min if and only if err == nil.
// If r returns an error having read at least min bytes, the error is dropped.

// ReadAtLeast方法从r中读取数据到buf中，直到至少读取了min个字节。 它返回拷贝的字节数和一个错误，如果读取的字节数少于min，则会返回错误。 如果没有读取任何字节，错误为EOF。 如果在读取的字节数少于min之后遇到EOF，则ReadAtLeast返回ErrUnexpectedEOF。 如果min大于buf的长度，ReadAtLeast返回ErrShortBuffer。 返回时，只有当err == nil时，n >= min。 如果r在读取了至少min个字节后返回一个错误，则会丢弃该错误。
func ReadAtLeast(r Reader, buf []byte, min int) (n int, err error) {
	if len(buf) < min {
		return 0, ErrShortBuffer
	}
	for n < min && err == nil {
		var nn int
		nn, err = r.Read(buf[n:])
		n += nn
	}
	if n >= min {
		err = nil
	} else if n > 0 && err == EOF {
		err = ErrUnexpectedEOF
	}
	return
}

// ReadFull reads exactly len(buf) bytes from r into buf.
// It returns the number of bytes copied and an error if fewer bytes were read.
// The error is EOF only if no bytes were read.
// If an EOF happens after reading some but not all the bytes,
// ReadFull returns ErrUnexpectedEOF.
// On return, n == len(buf) if and only if err == nil.
// If r returns an error having read at least len(buf) bytes, the error is dropped.

// ReadFull方法从r中精确地读取len(buf)个字节到buf中。 它返回拷贝的字节数和一个错误，如果读取的字节数少于len(buf)，则会返回错误。 如果没有读取任何字节，错误为EOF。 如果在读取了一些但不是全部字节后遇到EOF，则ReadFull返回ErrUnexpectedEOF。 返回时，只有当err == nil时，n == len(buf)。 如果r在读取了至少len(buf)个字节后返回一个错误，则会丢弃该错误。
func ReadFull(r Reader, buf []byte) (n int, err error) {
	return ReadAtLeast(r, buf, len(buf))
}

// CopyN copies n bytes (or until an error) from src to dst.
// It returns the number of bytes copied and the earliest
// error encountered while copying.
// On return, written == n if and only if err == nil.
//
// If dst implements the ReaderFrom interface,
// the copy is implemented using it.

// CopyN 方法从源（src）复制 n 个字节（或直到发生错误）到目标（dst）。它返回复制的字节数以及在复制过程中遇到的最早的错误。
// 在返回时，当且仅当 err nil 时，written n。
// 如果目标（dst）实现了 ReaderFrom 接口，则使用该接口来执行复制操作。
func CopyN(dst Writer, src Reader, n int64) (written int64, err error) {
	written, err = Copy(dst, LimitReader(src, n))
	if written == n {
		return n, nil
	}
	if written < n && err == nil {
		// src stopped early; must have been EOF.
		err = EOF
	}
	return
}

// Copy copies from src to dst until either EOF is reached
// on src or an error occurs. It returns the number of bytes
// copied and the first error encountered while copying, if any.
//
// A successful Copy returns err == nil, not err == EOF.
// Because Copy is defined to read from src until EOF, it does
// not treat an EOF from Read as an error to be reported.
//
// If src implements the WriterTo interface,
// the copy is implemented by calling src.WriteTo(dst).
// Otherwise, if dst implements the ReaderFrom interface,
// the copy is implemented by calling dst.ReadFrom(src).

// Copy 方法从源（src）复制数据到目标（dst），直到源（src）达到 EOF 或发生错误为止。它返回复制的字节数以及在复制过程中遇到的第一个错误（如果有的话）。
// 成功的 Copy 方法返回 err nil，而不是 err EOF。因为 Copy 方法定义为从源（src）读取直到 EOF，它不将从 Read 方法返回的 EOF 视为需要报告的错误。
// 如果源（src）实现了 WriterTo 接口，则复制操作通过调用 src.WriteTo(dst) 来实现。否则，如果目标（dst）实现了 ReaderFrom 接口，则复制操作通过调用 dst.ReadFrom(src) 来实现
func Copy(dst Writer, src Reader) (written int64, err error) {
	return copyBuffer(dst, src, nil)
}

// CopyBuffer is identical to Copy except that it stages through the
// provided buffer (if one is required) rather than allocating a
// temporary one. If buf is nil, one is allocated; otherwise if it has
// zero length, CopyBuffer panics.
//
// If either src implements WriterTo or dst implements ReaderFrom,
// buf will not be used to perform the copy.

// CopyBuffer 与 Copy 方法相同，只是它通过提供的缓冲区（如果需要）来进行暂存，而不是分配临时缓冲区。如果 buf 为 nil，则会分配一个缓冲区；如果 buf 长度为零，则 CopyBuffer 会引发 panic。
// 如果源（src）实现了 WriterTo 接口或目标（dst）实现了 ReaderFrom 接口，则不会使用缓冲区 buf 来执行复制操作
func CopyBuffer(dst Writer, src Reader, buf []byte) (written int64, err error) {
	if buf != nil && len(buf) == 0 {
		panic("empty buffer in CopyBuffer")
	}
	return copyBuffer(dst, src, buf)
}

// copyBuffer is the actual implementation of Copy and CopyBuffer.
// if buf is nil, one is allocated.

// copyBuffer 是 Copy 和 CopyBuffer 的实际实现。
// 如果 buf 为 nil，则会分配一个缓冲区
func copyBuffer(dst Writer, src Reader, buf []byte) (written int64, err error) {
	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.

	// 如果读取器具有 WriteTo 方法，则使用该方法来进行复制。
	// 这样可以避免分配和复制的开销。
	if wt, ok := src.(WriterTo); ok {
		return wt.WriteTo(dst)
	}
	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	// 类似地，如果给定的写入器实现了 ReadFrom 方法，就会使用该方法来进行数据的复制操作。
	if rt, ok := dst.(ReaderFrom); ok {
		return rt.ReadFrom(src)
	}
	if buf == nil {
		size := 32 * 1024
		if l, ok := src.(*LimitedReader); ok && int64(size) > l.N {
			if l.N < 1 {
				size = 1
			} else {
				size = int(l.N)
			}
		}
		buf = make([]byte, size)
	}
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

// LimitReader returns a Reader that reads from r
// but stops with EOF after n bytes.
// The underlying implementation is a *LimitedReader.

// LimitReader 返回一个读取器，该读取器从 r 中读取数据，但在读取 n 字节后停止并返回 EOF。
// 其底层实现是一个 *LimitedReader。
func LimitReader(r Reader, n int64) Reader { return &LimitedReader{r, n} }

// A LimitedReader reads from R but limits the amount of
// data returned to just N bytes. Each call to Read
// updates N to reflect the new amount remaining.
// Read returns EOF when N <= 0 or when the underlying R returns EOF.

// LimitedReader 是一个从 R 读取数据的读取器，但是限制了返回的数据量为 N 字节。
// 每次调用 Read 都会更新 N 的值以反映剩余的数据量。
// 当 N <= 0 或者底层的 R 返回 EOF 时，Read 返回 EOF。
type LimitedReader struct {
	R Reader // underlying reader | 底层的读取器
	N int64  // max bytes remaining | 限制的最大剩余字节数
}

func (l *LimitedReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.R.Read(p)
	l.N -= int64(n)
	return
}

// NewSectionReader returns a SectionReader that reads from r
// starting at offset off and stops with EOF after n bytes.
func NewSectionReader(r ReaderAt, off int64, n int64) *SectionReader {
	var remaining int64
	const maxint64 = 1<<63 - 1
	if off <= maxint64-n {
		remaining = n + off
	} else {
		// Overflow, with no way to return error.
		// Assume we can read up to an offset of 1<<63 - 1.
		remaining = maxint64
	}
	return &SectionReader{r, off, off, remaining}
}

// SectionReader implements Read, Seek, and ReadAt on a section
// of an underlying ReaderAt.
type SectionReader struct {
	r     ReaderAt
	base  int64
	off   int64
	limit int64
}

func (s *SectionReader) Read(p []byte) (n int, err error) {
	if s.off >= s.limit {
		return 0, EOF
	}
	if max := s.limit - s.off; int64(len(p)) > max {
		p = p[0:max]
	}
	n, err = s.r.ReadAt(p, s.off)
	s.off += int64(n)
	return
}

var errWhence = errors.New("Seek: invalid whence")
var errOffset = errors.New("Seek: invalid offset")

func (s *SectionReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	default:
		return 0, errWhence
	case SeekStart:
		offset += s.base
	case SeekCurrent:
		offset += s.off
	case SeekEnd:
		offset += s.limit
	}
	if offset < s.base {
		return 0, errOffset
	}
	s.off = offset
	return offset - s.base, nil
}

func (s *SectionReader) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 || off >= s.limit-s.base {
		return 0, EOF
	}
	off += s.base
	if max := s.limit - off; int64(len(p)) > max {
		p = p[0:max]
		n, err = s.r.ReadAt(p, off)
		if err == nil {
			err = EOF
		}
		return n, err
	}
	return s.r.ReadAt(p, off)
}

// Size returns the size of the section in bytes.
func (s *SectionReader) Size() int64 { return s.limit - s.base }

// An OffsetWriter maps writes at offset base to offset base+off in the underlying writer.
type OffsetWriter struct {
	w    WriterAt
	base int64 // the original offset
	off  int64 // the current offset
}

// NewOffsetWriter returns an OffsetWriter that writes to w
// starting at offset off.
func NewOffsetWriter(w WriterAt, off int64) *OffsetWriter {
	return &OffsetWriter{w, off, off}
}

func (o *OffsetWriter) Write(p []byte) (n int, err error) {
	n, err = o.w.WriteAt(p, o.off)
	o.off += int64(n)
	return
}

func (o *OffsetWriter) WriteAt(p []byte, off int64) (n int, err error) {
	off += o.base
	return o.w.WriteAt(p, off)
}

func (o *OffsetWriter) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	default:
		return 0, errWhence
	case SeekStart:
		offset += o.base
	case SeekCurrent:
		offset += o.off
	}
	if offset < o.base {
		return 0, errOffset
	}
	o.off = offset
	return offset - o.base, nil
}

// TeeReader returns a Reader that writes to w what it reads from r.
// All reads from r performed through it are matched with
// corresponding writes to w. There is no internal buffering -
// the write must complete before the read completes.
// Any error encountered while writing is reported as a read error.
func TeeReader(r Reader, w Writer) Reader {
	return &teeReader{r, w}
}

type teeReader struct {
	r Reader
	w Writer
}

func (t *teeReader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		if n, err := t.w.Write(p[:n]); err != nil {
			return n, err
		}
	}
	return
}

// Discard is a Writer on which all Write calls succeed
// without doing anything.

// Discard 是一个 Writer，所有 Write 调用都会成功，但不做任何东西。
var Discard Writer = discard{}

type discard struct{}

// discard implements ReaderFrom as an optimization so Copy to
// io.Discard can avoid doing unnecessary work.
var _ ReaderFrom = discard{}

func (discard) Write(p []byte) (int, error) {
	return len(p), nil
}

func (discard) WriteString(s string) (int, error) {
	return len(s), nil
}

var blackHolePool = sync.Pool{
	New: func() any {
		b := make([]byte, 8192)
		return &b
	},
}

func (discard) ReadFrom(r Reader) (n int64, err error) {
	bufp := blackHolePool.Get().(*[]byte)
	readSize := 0
	for {
		readSize, err = r.Read(*bufp)
		n += int64(readSize)
		if err != nil {
			blackHolePool.Put(bufp)
			if err == EOF {
				return n, nil
			}
			return
		}
	}
}

// NopCloser returns a ReadCloser with a no-op Close method wrapping
// the provided Reader r.
// If r implements WriterTo, the returned ReadCloser will implement WriterTo
// by forwarding calls to r.
func NopCloser(r Reader) ReadCloser {
	if _, ok := r.(WriterTo); ok {
		return nopCloserWriterTo{r}
	}
	return nopCloser{r}
}

type nopCloser struct {
	Reader
}

func (nopCloser) Close() error { return nil }

type nopCloserWriterTo struct {
	Reader
}

func (nopCloserWriterTo) Close() error { return nil }

func (c nopCloserWriterTo) WriteTo(w Writer) (n int64, err error) {
	return c.Reader.(WriterTo).WriteTo(w)
}

// ReadAll reads from r until an error or EOF and returns the data it read.
// A successful call returns err == nil, not err == EOF. Because ReadAll is
// defined to read from src until EOF, it does not treat an EOF from Read
// as an error to be reported.

// ReadAll 方法从读取器（r）读取数据，直到遇到错误或 EOF，并返回读取到的数据。
// 成功的调用返回的 err 值为 nil，而不是 EOF。因为 ReadAll 方法定义为从源（r）读取直到 EOF，它不将从 Read 方法返回的 EOF 视为需要报告的错误。
func ReadAll(r Reader) ([]byte, error) {
	b := make([]byte, 0, 512)
	for {
		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == EOF {
				err = nil
			}
			return b, err
		}
	}
}
