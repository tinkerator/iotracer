// Package iotracer maintains a circular buffer of trace entries where
// traces hold flat 64 bit mask and value pairs time stamped.
package iotracer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// datum holds a single trace entry.
type Sample struct {
	// When holds the time the sample was logged.
	When time.Time
	// Mask and Value capture the data recorded in this sample.
	Mask, Value uint64
}

// Trace holds a tracer.
type Trace struct {
	// app names the application that is making the trace.
	app string

	// mu protects the following entries.
	mu sync.Mutex

	// samples holds a circular buffer of data
	samples []Sample

	// maxSamples indicates the total number of samples held in
	// the buffer
	maxSamples uint

	// cursor is a monotonically incremented counter indicating
	// where (% .maxSamples) the next datum will be stored in
	// .array.
	cursor uint

	// fullMask holds the union of all mask values so far.
	fullMask uint64

	// labels capture the preferred labels for each numbered
	// signal. The default labels are sig<n> where n is the [0,63)
	// index of the traced bit value.
	labels map[int]string
}

// NewTrace allocates a new tracer, capable of storing up to samples
// of recent samples.
func NewTrace(app string, samples uint) *Trace {
	if samples == 0 {
		return nil
	}
	if app == "" {
		app = "iotracer"
	}
	return &Trace{
		app:        app,
		samples:    make([]Sample, samples),
		maxSamples: samples,
		labels:     make(map[int]string),
	}
}

// ErrInvalidSignalIndex is returned if an attempt is made to
// reference an impossible signal bit.
var ErrInvalidSignalIndex = errors.New("invalid signal index, want [0,64)")

// Label names a specific signal offset with a text label. If
// label="", the label reverts to its default value: "sig#".
func (t *Trace) Label(sig int, label string) error {
	if t == nil || sig < 0 || sig >= 64 {
		return ErrInvalidSignalIndex
	}
	if label == "" {
		delete(t.labels, sig)
	} else {
		t.labels[sig] = label
	}
	return nil
}

// SampleAt appends a new snapshot to the trace at the specified time
// if the sample differs from the previously recorded one and the
// now value is after the most recently recorded sample.
func (t *Trace) SampleAt(now time.Time, mask, value uint64) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	j := (t.cursor + t.maxSamples - 1) % t.maxSamples
	if s := t.samples[j]; t.cursor != 0 && ((s.Mask == mask && s.Value == value) || !now.After(s.When)) {
		return
	}
	datum := &t.samples[t.cursor%t.maxSamples]
	t.fullMask |= mask
	t.cursor++
	datum.When = now
	datum.Mask = mask
	datum.Value = value
}

// Sample appends a new snapshot to the trace, unless the new
// snapshot exactly matches the previously captured value.
func (t *Trace) Sample(mask, value uint64) {
	t.SampleAt(time.Now(), mask, value)
}

// Output a section named such containing content with an end.
func (t *Trace) vcdSection(w io.Writer, section, content string, oneLine bool) {
	if oneLine {
		fmt.Fprintf(w, "$%s %s $end\n", section, content)
	} else {
		fmt.Fprintf(w, "$%s\n\t%s\n$end\n", section, content)
	}
}

// ErrNoTraceData indicates the current trace contains no data.
var ErrNoTraceData = errors.New("no trace data")

type signal struct {
	mask uint64
	ch   string
	lab  string
}

// VCD generates a Value Change Dump from the trace recorded so far.
// The function starts by making a snapshot of the current trace.
func (t *Trace) VCD(tScale time.Duration) (io.Reader, error) {
	if t == nil {
		return nil, ErrNoTraceData
	}

	// Lock to make a thread safe copy of the data.
	t.mu.Lock()
	fullMask := t.fullMask
	app := t.app
	samples := t.cursor
	cursor := t.cursor
	if cursor > t.maxSamples {
		samples = t.maxSamples
	}
	working := make([]Sample, samples)
	start := (cursor + t.maxSamples - samples) % t.maxSamples

	if samples != cursor {
		copy(working[:samples-start], t.samples[start:])
		copy(working[samples-start:], t.samples[:start])
	} else {
		copy(working[:samples], t.samples[:samples])
	}
	var sigs []signal
	for i, j, bit := 0, 0, uint64(1); bit != 0 && bit < fullMask; i, bit = i+1, bit<<1 {
		if fullMask&bit != 0 {
			lab := t.labels[i]
			if lab == "" {
				lab = fmt.Sprintf("sig%d", i)
			}
			sig := signal{
				mask: bit,
				ch:   fmt.Sprintf("%c", 33+j),
				lab:  lab,
			}
			j++
			sigs = append(sigs, sig)
		}
	}
	t.mu.Unlock()

	w := &bytes.Buffer{}

	from := working[0].When

	t.vcdSection(w, "date", from.Format(time.ANSIC), false)
	t.vcdSection(w, "version", app, false)
	t.vcdSection(w, "timescale", fmt.Sprintf("%v", tScale), false)
	t.vcdSection(w, "scope", "module ports", true)

	for _, sig := range sigs {
		t.vcdSection(w, "var", fmt.Sprintf("wire 1 %s %s", sig.ch, sig.lab), true)
	}

	fmt.Fprint(w, "$upscope $end\n")
	fmt.Fprint(w, "$enddefinitions $end\n")
	fmt.Fprint(w, "#0\n")
	fmt.Fprint(w, "$dumpvars\n")

	var lastVal, lastMask uint64
	var lastStamp = 0
	for i := uint(0); i < samples; i++ {
		s := working[i]
		delta := lastVal ^ s.Value
		dMask := lastMask ^ s.Mask
		if i == 0 || dMask|delta != 0 {
			// Something has changed, so we need to include it in the dump file.
			stamp := int(s.When.Sub(from) / tScale)
			if stamp != lastStamp {
				fmt.Fprintf(w, "#%d\n", stamp)
				lastStamp = stamp
			}
			for _, sig := range sigs {
				if sig.mask&s.Mask == 0 {
					if i == 0 || (sig.mask&dMask) != 0 {
						fmt.Fprintf(w, "x%s\n", sig.ch)
					}
				} else {
					if i == 0 || (sig.mask&delta) != 0 {
						v := 0
						if sig.mask&s.Value != 0 {
							v = 1
						}
						fmt.Fprintf(w, "%d%s\n", v, sig.ch)
					}
				}
			}
		}
		lastVal = s.Value
		lastMask = s.Mask
	}

	return w, nil
}
