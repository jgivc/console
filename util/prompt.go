package util

import (
	"errors"
	"io"
	"os"
	"regexp"
	"time"
)

var ErrNoPromptFound = errors.New("no prompt found")

type TimeoutReader interface {
	io.ReadCloser
	SetReadTimeout(t time.Duration)
}

type promptReader struct {
	deadLine    time.Time
	err         error
	matchLength int
	buf         Buffer
	reg         *regexp.Regexp
	reader      TimeoutReader
	returnOnly  bool
}

func (r *promptReader) SetPromptPattern(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	r.reg = re

	return nil
}

func (r *promptReader) SetDeadLine(deadLine time.Time) {
	r.deadLine = deadLine
}

func (r *promptReader) Reset() {
	r.buf.Reset()
	r.err = nil
}

/*
The read method reads from the underlying TimeoutReader until it finds a prompt.
If the prompt is found, io.EOF will be returned. If deadline is reached ErrNoPromptFound will be retirned.
*/
func (r *promptReader) Read(p []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	if r.returnOnly {
		n := copy(p, r.buf.Bytes())
		if n == r.buf.Len() {
			r.returnOnly = false
			r.buf.Reset()

			r.err = io.EOF
		} else {
			r.buf.Shift(n)
		}

		return n, nil
	}

	i := 0

	for {
		i++
		n, err := r.buf.ReadFrom(r.reader)
		if err != nil {
			if errors.Is(err, io.EOF) && n < 1 {
				r.err = ErrNoPromptFound
				return 0, r.err
			}

			if errors.Is(err, os.ErrDeadlineExceeded) {
				if time.Now().After(r.deadLine) {
					r.err = ErrNoPromptFound
					return 0, r.err
				}
				continue
			}
		}

		if loc := r.reg.FindIndex(r.buf.Bytes()); loc != nil {
			n2 := copy(p, r.buf.Bytes())
			if n2 != r.buf.Len() {
				r.returnOnly = true
				r.buf.Shift(n2)
				return n2, nil
			}

			r.err = io.EOF
			return n2, nil
		} else if errors.Is(err, io.EOF) {
			r.err = ErrNoPromptFound
			return 0, r.err
		}

		if r.buf.Len() > r.matchLength {
			l := r.buf.Len() - r.matchLength
			n3 := copy(p, r.buf.Bytes()[:l])
			r.buf.Shift(n3)

			return n3, nil
		}

		if time.Now().After(r.deadLine) {
			r.err = ErrNoPromptFound
			return 0, r.err
		}
	}
}

func NewPromptReader(reader TimeoutReader, buffSize int, matchLength int) *promptReader {
	return &promptReader{
		matchLength: matchLength,
		reader:      reader,
	}
}
