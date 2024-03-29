// Program sample outputs a sample VCD dump.
package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"zappem.net/pub/io/iotracer"
)

const layout = ""

func main() {
	tr := iotracer.NewTrace("sample", 100)
	tr2 := iotracer.NewTrace("other", 30)
	ch, err := tr.Watch(3, 100)
	if err != nil {
		log.Fatalf("unable to watch signal 3: %v", err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for ev := range ch {
			fmt.Fprintf(os.Stderr, "sample.ports.sig3 @ %s = %v\n", ev.When.Format("2006-01-02 15:04:05.000000000"), ev.On)
		}
	}()

	var mask uint64
	t := time.Now()

	for i := uint64(0); i < 33; i++ {
		mask |= i
		t = t.Add(time.Duration(i) * 100 * time.Nanosecond)
		tr.SampleAt(t, mask, i)
		tr2.SampleAt(t.Add(10000), mask, i)
	}
	for i := uint64(0); i < 33; i++ {
		mask ^= i
		t = t.Add(time.Duration(i) * 100 * time.Nanosecond)
		tr.SampleAt(t, mask, i)
	}
	tr.Cancel(ch)
	wg.Wait()

	// merge together 3 bits.
	if err := tr2.Alias(1, 0, 2, "octo"); err != nil {
		log.Fatalf("unable to alias 3 bits of tr2: %v", err)
	}

	b, err := iotracer.ExportVCD("top", 100*time.Nanosecond, tr, tr2)
	if err != nil {
		log.Fatalf("failed to dump trace: %v", err)
	}
	for line := range b {
		fmt.Println(line)
	}
}
