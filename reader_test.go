package console

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

type fr struct {
	buf   *bytes.Buffer
	pause time.Duration
}

func (r *fr) Read(p []byte) (int, error) {
	if r.pause > 0 {
		time.Sleep(r.pause)
	}

	return r.buf.Read(p)
}

func TestTReader(t *testing.T) {
	data := []struct {
		data    string
		timeout time.Duration
	}{
		{
			"aaaaaaaaaaaa;lk ;;lk;lk43r43r40jrf 4rf45f54 f4f",
			0,
		},
		{
			"aaaaaaaaaaaa;lk ;;lk;lk43r43r40jrf 4rf45f54 f4f 43rr43 43543 5435 43\n435435 3465 43",
			30 * time.Millisecond,
		},
		{
			"oi043r43 r435095lkr4m3tv54tv54 t54yt;l54-06mt 4566v 6546b57577n765765 7657657b65765765756765polkp;56;l;l;lk;lk;lk;k;5657bv6576576577657657b657",
			80 * time.Millisecond,
		},
	}

	sb := &strings.Builder{}

	for _, d := range data {
		r := &fr{buf: bytes.NewBufferString(d.data), pause: d.timeout}
		tr := newTimeoutReader(r)

		b := make([]byte, 100)

		if d.timeout > 0 {
			start := time.Now()

			tm := d.timeout - 5*time.Millisecond

			tr.SetTimeout(tm)
			_, err := tr.Read(b)
			if err == nil || err != errorRreadTimeout {
				t.Fatalf("Timeout must be")
			}

			if tm.Milliseconds() != time.Since(start).Milliseconds() {
				t.Fatalf("Want timeout %d, but got %d\n", tm.Milliseconds(), time.Since(start).Milliseconds())
			}

			tr.SetTimeout(d.timeout + 20*time.Millisecond)
		}

		sb.Reset()

		for {
			n, err := tr.Read(b)
			if err != nil {
				if err == io.EOF {
					break
				}

				t.Fatal(err)
			}

			if err := sCopy(sb, b[:n]); err != nil {
				t.Fatal(err)
			}

		}

		if d.data != sb.String() {
			t.Fatalf("different strings")
		}

	}
}
