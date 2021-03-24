package utilsio

import (
	"io"
	"time"
)

type readerData struct {
	bytes []byte
	err   error
}

// ReadWithTimeout is reading bytes chunks from reader and combine them togheter in one bytes slice
// Timeout is for stream reader which wait for given time to receive next stream of bytes
// End of operation is when either timeout has come or any error has been thrown
func ReadWithTimeout(r io.Reader, timeout time.Duration) ([]byte, error) {
	bytesCh := make(chan readerData)

	go func() {
		bytesChunkCh := make(chan readerData)
		bytes := make([]byte, 0)
		tch := time.NewTimer(timeout)
		defer tch.Stop()

		var err error
		isRunning := true
		for {
			if !isRunning {
				break
			}
			go readBytesChunk(bytesChunkCh, r)
			// Reset timer
			tch.Reset(timeout)

			select {
			case bmCh, more := <-bytesChunkCh:
				bytes = append(bytes, bmCh.bytes...)

				if bmCh.err != nil {
					err = bmCh.err
					isRunning = false
					break
				}
				if !more {
					isRunning = false
					break
				}
			case <-tch.C:
				tch.Stop()
				isRunning = false
				break
			}
		}

		bytesCh <- readerData{
			bytes: bytes,
			err:   err,
		}
	}()

	done := <-bytesCh

	return done.bytes, done.err
}

func readBytesChunk(rdCh chan<- readerData, reader io.Reader) {
	b := make([]byte, 1024)
	l, err := reader.Read(b)
	filled := b[:l]
	if err != nil {
		rdCh <- readerData{
			bytes: filled,
			err:   err,
		}
		return
	}

	rdCh <- readerData{
		bytes: filled,
		err:   nil,
	}
}
