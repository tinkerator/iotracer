// Program sample outputs a sample VCD dump.
package main

import (
	"io"
	"log"
	"os"
	"time"

	"zappem.net/pub/io/iotracer"
)

func main() {
	tr := iotracer.NewTrace("sample", 100)

	var mask uint64
	t := time.Now()

	for i := uint64(0); i < 33; i++ {
		mask |= i
		t = t.Add(time.Duration(i) * 100 * time.Nanosecond)
		tr.SampleAt(t, mask, i)
	}
	for i := uint64(0); i < 33; i++ {
		mask ^= i
		t = t.Add(time.Duration(i) * 100 * time.Nanosecond)
		tr.SampleAt(t, mask, i)
	}

	b, err := tr.VCD(100 * time.Nanosecond)
	if err != nil {
		log.Fatalf("failed to dump trace: %v", err)
	}
	io.Copy(os.Stdout, b)
}
