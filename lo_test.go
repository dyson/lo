// Copyright 2017 Dyson Simmons. All rights reserved.

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lo_test

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/dyson/lo"
)

const (
	Rdate         = `[0-9][0-9][0-9][0-9]/[0-9][0-9]/[0-9][0-9]`
	Rtime         = `[0-9][0-9]:[0-9][0-9]:[0-9][0-9]`
	Rmicroseconds = `\.[0-9][0-9][0-9][0-9][0-9][0-9]`
	Rline         = `(58|80):` // must update if the calls to l.Printf / l.Print below move
	Rlongfile     = `.*/[A-Za-z0-9_\-]+\.go:` + Rline
	Rshortfile    = `[A-Za-z0-9_\-]+\.go:` + Rline
)

type tester struct {
	flag    int
	prefix  string
	pattern string // regexp that log output must match; we add ^ and expected_text$ always
}

var tests = []tester{
	// individual pieces:
	{0, "", ""},
	{0, "XXX", "XXX"},
	{lo.Ldate, "", Rdate + " "},
	{lo.Ltime, "", Rtime + " "},
	{lo.Ltime | lo.Lmicroseconds, "", Rtime + Rmicroseconds + " "},
	{lo.Lmicroseconds, "", Rtime + Rmicroseconds + " "}, // microsec implies time
	{lo.Llongfile, "", Rlongfile + " "},
	{lo.Lshortfile, "", Rshortfile + " "},
	{lo.Llongfile | lo.Lshortfile, "", Rshortfile + " "}, // shortfile overrides longfile
	// everything at once:
	{lo.Ldate | lo.Ltime | lo.Lmicroseconds | lo.Llongfile, "XXX", "XXX" + Rdate + " " + Rtime + Rmicroseconds + " " + Rlongfile + " "},
	{lo.Ldate | lo.Ltime | lo.Lmicroseconds | lo.Lshortfile, "XXX", "XXX" + Rdate + " " + Rtime + Rmicroseconds + " " + Rshortfile + " "},
}

// Test using using Printf("hello %d world", 23)
func testPrint(t *testing.T, flag int, prefix string, pattern string, logLevel int) {
	buf := new(bytes.Buffer)
	l := lo.New(buf, prefix, flag)
	l.SetLevel(logLevel)

	l.Printf("hello %d world", 23)

	if l.Level() != lo.LevelDebug {
		l.Printf("debug: hello %d world", 23)
	}

	line := buf.String()
	line = line[0 : len(line)-1]

	buf.Reset()

	pattern1 := "^" + pattern + "INFO hello 23 world$"
	matched, err := regexp.MatchString(pattern1, line)

	if err != nil {
		t.Fatal("pattern did not compile:", err)
	}
	if !matched {
		t.Errorf("log output should match %q is %q", pattern1, line)
	}

	if l.Level() == lo.LevelDebug {
		l.Printf("debug: hello %d world", 23)

		line := buf.String()
		line = line[0 : len(line)-1]

		pattern2 := "^" + pattern + "DEBUG hello 23 world$"
		matched, err := regexp.MatchString(pattern2, line)

		if err != nil {
			t.Fatal("pattern did not compile:", err)
		}
		if !matched {
			t.Errorf("log output should match %q is %q", pattern2, line)
		}
	}

	l.SetOutput(os.Stderr)
}

func TestAll(t *testing.T) {
	for _, testcase := range tests {
		testPrint(t, testcase.flag, testcase.prefix, testcase.pattern, lo.LevelInfo)
		testPrint(t, testcase.flag, testcase.prefix, testcase.pattern, lo.LevelDebug)
	}
}

func TestFlagAndPrefixSetting(t *testing.T) {
	buf := new(bytes.Buffer)
	l := lo.New(buf, "Test:", lo.LstdFlags)
	f := l.Flags()
	if f != lo.LstdFlags {
		t.Errorf("Flags 1: expected %x got %x", lo.LstdFlags, f)
	}
	l.SetFlags(f | lo.Lmicroseconds)
	f = l.Flags()
	if f != lo.LstdFlags|lo.Lmicroseconds {
		t.Errorf("Flags 2: expected %x got %x", lo.LstdFlags|lo.Lmicroseconds, f)
	}
	p := l.Prefix()
	if p != "Test:" {
		t.Errorf(`Prefix: expected "Test:" got %q`, p)
	}
	l.SetPrefix("Reality:")
	p = l.Prefix()
	if p != "Reality:" {
		t.Errorf(`Prefix: expected "Reality:" got %q`, p)
	}
	// Verify a log message looks right, with our prefix and microseconds present.
	l.SetLevel(lo.LevelDebug)
	l.Printf("debug:hello")
	line := buf.Bytes()
	pattern := "^Reality:" + Rdate + " " + Rtime + Rmicroseconds + " DEBUG hello\n"
	matched, err := regexp.Match(pattern, line)
	if err != nil {
		t.Fatalf("pattern %q did not compile: %s", pattern, err)
	}
	if !matched {
		t.Errorf("log output should match %q is %q", pattern, line)
	}
}

func TestUTCFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	l := lo.New(buf, "Test:", lo.LstdFlags)
	l.SetFlags(lo.Ldate | lo.Ltime | lo.LUTC)
	// Verify a log message looks right in the right time zone. Quantize to the second only.
	now := time.Now().UTC()
	l.Printf("test")
	want := fmt.Sprintf("Test:%d/%.2d/%.2d %.2d:%.2d:%.2d INFO test\n",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	got := buf.String()
	if got == want {
		return
	}
	// It's possible we crossed a second boundary between getting now and logging,
	// so add a second and try again. This should very nearly always work.
	now = now.Add(time.Second)
	want = fmt.Sprintf("Test:%d/%.2d/%.2d %.2d:%.2d:%.2d INFO test\n",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	if got == want {
		return
	}
	t.Errorf("got %q; want %q", got, want)
}

func TestEmptyPrintCreatesLine(t *testing.T) {
	buf := new(bytes.Buffer)
	l := lo.New(buf, "Header:", lo.LstdFlags)
	l.Printf("")
	l.Printf("test")
	output := buf.String()
	if n := strings.Count(output, "Header"); n != 2 {
		t.Errorf("expected 2 headers, got %d", n)
	}
	if n := strings.Count(output, "\n"); n != 2 {
		t.Errorf("expected 2 lines, got %d", n)
	}
}

func TestLevelNone(t *testing.T) {
	buf := new(bytes.Buffer)
	l := lo.New(buf, "", 0)
	l.SetLevel(lo.LevelNone)
	l.Printf("test")
	line := buf.String()
	if line != "" {
		t.Errorf("expected no output, got %q", line)
	}
}

func BenchmarkPrintfNone(b *testing.B) {
	const testString = "test"
	var buf bytes.Buffer
	l := lo.New(&buf, "", lo.LstdFlags)
	l.SetLevel(lo.LevelNone)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		l.Printf(testString)
	}
}

func BenchmarkPrintfNoneNoFlags(b *testing.B) {
	const testString = "test"
	var buf bytes.Buffer
	l := lo.New(&buf, "", 0)
	l.SetLevel(lo.LevelNone)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		l.Printf(testString)
	}
}

func BenchmarkPrintfInfo(b *testing.B) {
	const testString = "test"
	var buf bytes.Buffer
	l := lo.New(&buf, "", lo.LstdFlags)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		l.Printf(testString)
	}
}

func BenchmarkPrintfInfoNoFlags(b *testing.B) {
	const testString = "test"
	var buf bytes.Buffer
	l := lo.New(&buf, "", 0)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		l.Printf(testString)
	}
}

func BenchmarkPrintfDebug(b *testing.B) {
	const testString = "debug: test"
	var buf bytes.Buffer
	l := lo.New(&buf, "", lo.LstdFlags)
	l.SetLevel(lo.LevelDebug)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		l.Printf(testString)
	}
}

func BenchmarkPrintfDebugNoFlags(b *testing.B) {
	const testString = "debug: test"
	var buf bytes.Buffer
	l := lo.New(&buf, "", 0)
	l.SetLevel(lo.LevelDebug)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		l.Printf(testString)
	}
}
