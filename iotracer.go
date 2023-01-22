// Package iotracer maintains a circular buffer of trace entries where
// traces hold flat 64 bit mask and value pairs time stamped.
package iotracer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
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

type Event struct {
	// When the event was generated.
	When time.Time
	// What the value of the signal was at that time.
	On bool
}

// Trace holds a tracer.
type Trace struct {
	// app names the application that is making the trace.
	app string

	// module names the signal root. If this is empty, it
	// is reported as "ports".
	module string

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

	// changes hold channels to write to when tracked IO values
	// change.
	changes map[uint64][]chan Event
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
		changes:    make(map[uint64][]chan Event),
	}
}

// Module sets the trace module name. This is output in the VCD dump
// to group the signals. The default, which can be recovered by
// setting the value to "" is "ports".
func (t *Trace) Module(name string) {
	if t == nil {
		return
	}
	t.module = name
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

// Watch opens an Event channel to watch for changes in value of a
// specified signal.
func (t *Trace) Watch(sig, depth int) (<-chan Event, error) {
	if t == nil || sig < 0 || sig >= 64 {
		return nil, ErrInvalidSignalIndex
	}
	mask := uint64(1) << uint(sig)
	ch := make(chan Event, depth)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.changes[mask] = append(t.changes[mask], ch)
	return ch, nil
}

// ErrUnknownWatcher is used to signal that a channel was not
// recognized and has not been closed.
var ErrUnknownWatcher = errors.New("unknown watcher")

// Cancel cancels a Watch() channel. By cancel, we mean close. Note
// while this function is being called other events may be channeled.
func (t *Trace) Cancel(ch <-chan Event) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for mask, chs := range t.changes {
		for i, comm := range chs {
			if (<-chan Event)(comm) == ch {
				t.changes[mask] = append(chs[:i], chs[i+1:]...)
				close(comm)
				return nil
			}
		}
	}
	return ErrUnknownWatcher
}

// SampleAt appends a new snapshot to the trace at the specified time
// if the sample differs from the previously recorded one and the
// now value is after the most recently recorded sample. The trace
// entries may cause Watch() events to be signaled.
func (t *Trace) SampleAt(now time.Time, mask, value uint64) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	j := (t.cursor + t.maxSamples - 1) % t.maxSamples
	s := t.samples[j]
	if t.cursor != 0 && ((s.Mask == mask && s.Value == value) || !now.After(s.When)) {
		return
	}

	datum := &t.samples[t.cursor%t.maxSamples]
	datum.When = now
	datum.Mask = mask
	datum.Value = value

	if delta := (s.Mask & s.Value) ^ (mask & value); t.cursor == 0 || delta != 0 {
		for m, evs := range t.changes {
			if m&delta == 0 {
				continue
			}
			on := m&mask&value != 0
			for _, ch := range evs {
				select {
				case ch <- Event{When: now, On: on}:
				default:
					// Never waits.
				}
			}
		}
	}

	t.fullMask |= mask
	t.cursor++
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

// signal is used for VCD signal identification.
type signal struct {
	mask uint64
	ch   string
	lab  string
}

// keyOf represents the number j in a unique VCD preferred format.
func keyOf(j int) string {
	var cs []string
	const digit = 127 - 33
	const base = 33
	for loop := true; loop; loop = j != 0 {
		c := j % digit
		cs = append(cs, fmt.Sprintf("%c", base+c))
		j /= digit
	}
	return strings.Join(cs, "")
}

// VCD generates a Value Change Dump from the trace recorded so far.
// The function starts by making a snapshot of the current trace.
func (t *Trace) VCD(tScale time.Duration) (io.Reader, error) {
	if t == nil {
		return nil, ErrNoTraceData
	}

	// Lock to make a thread safe copy of the data.
	t.mu.Lock()
	module := "ports"
	if t.module != "" {
		module = t.module
	}
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
				ch:   keyOf(j),
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
	t.vcdSection(w, "scope", fmt.Sprintf("module %s", module), true)

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
