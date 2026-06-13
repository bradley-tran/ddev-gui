package backend

import (
	"bytes"
	"testing"
	"reflect"
)

func scanLinesOrCRIterative(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i, b := range data {
		if b == '\n' {
			return i + 1, data[:i], nil
		}
		if b == '\r' {
			// \r\n counts as a single line break
			if i+1 < len(data) && data[i+1] == '\n' {
				return i + 2, data[:i], nil
			}
			return i + 1, data[:i], nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}


func TestScanLinesOrCREquivalence(t *testing.T) {
	cases := []struct{
		name string
		data []byte
		atEOF bool
	}{
		{"empty", []byte(""), true},
		{"empty not at eof", []byte(""), false},
		{"no break", []byte("hello"), false},
		{"no break at eof", []byte("hello"), true},
		{"newline", []byte("hello\nworld"), false},
		{"carriage", []byte("hello\rworld"), false},
		{"crlf", []byte("hello\r\nworld"), false},
		{"cr crlf", []byte("hello\r\r\nworld"), false},
		{"crlf trailing", []byte("hello\r\n"), false},
		{"cr trailing", []byte("hello\r"), false},
		{"cr trailing eof", []byte("hello\r"), true},
		{"only cr", []byte("\r"), false},
		{"only lf", []byte("\n"), false},
		{"only crlf", []byte("\r\n"), false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			adv1, tok1, err1 := scanLinesOrCRIterative(c.data, c.atEOF)
			adv2, tok2, err2 := scanLinesOrCR(c.data, c.atEOF)

			if adv1 != adv2 || !reflect.DeepEqual(tok1, tok2) || err1 != err2 {
				t.Errorf("Mismatch for %q atEOF=%v:\nIter: adv=%d, tok=%q, err=%v\nOpt: adv=%d, tok=%q, err=%v", c.data, c.atEOF, adv1, tok1, err1, adv2, tok2, err2)
			}
		})
	}
}

func BenchmarkScanLinesOrCR_Iterative(b *testing.B) {
	data := []byte("This is a long string that does not contain any newlines or carriage returns until the very end.\r\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanLinesOrCRIterative(data, false)
	}
}

func BenchmarkScanLinesOrCR_Optimized(b *testing.B) {
	data := []byte("This is a long string that does not contain any newlines or carriage returns until the very end.\r\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanLinesOrCR(data, false)
	}
}

func BenchmarkScanLinesOrCR_Iterative_LongLine(b *testing.B) {
	data := append(bytes.Repeat([]byte("a"), 64*1024), '\n')
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanLinesOrCRIterative(data, false)
	}
}

func BenchmarkScanLinesOrCR_Optimized_LongLine(b *testing.B) {
	data := append(bytes.Repeat([]byte("a"), 64*1024), '\n')
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanLinesOrCR(data, false)
	}
}
