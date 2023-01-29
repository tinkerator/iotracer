// Package iotracer maintains a circular buffer of trace entries where
// traces hold flat 64 bit mask and value pairs time stamped.
package iotracer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"
)

// Sample holds a single trace entry.
type Sample struct {
	// When holds the time the sample was logged.
	When time.Time
	// Mask and Value capture the data recorded in this sample.
	Mask, Value uint64
}

// Event indicates when a signal value changed and what its new value
// was at that time.
type Event struct {
	// When the event was generated.
	When time.Time
	// What the value of the signal was at that time.
	On bool
}

// Trace holds a tracer.
type Trace struct {
	// app names the subsystem that is making the trace.
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
	// where (% .maxSamples) the next Sample will be stored in
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
func vcdSection(ch chan<- string, section, content string, oneLine bool) {
	if oneLine {
		ch <- fmt.Sprintf("$%s %s $end", section, content)
	} else {
		ch <- fmt.Sprint("$", section)
		ch <- fmt.Sprint("\t", content)
		ch <- "$end"
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

// VCDDetail holds everything needed to produce a VCD dump for a
// single trace.
type VCDDetail struct {
	app      string
	module   string
	fullMask uint64
	samples  uint
	working  []Sample
	start    uint
	sigs     []signal
}

// cacheVCDDetail snapshots a Trace for the purpose of generating a
// VCD trace from it. The provided index is the next available VCD
// signal index value and the returned int is the next available one.
func (t *Trace) cacheVCDDetail(index int) (*VCDDetail, int) {
	// Lock to make a thread safe copy of the data.
	t.mu.Lock()
	defer t.mu.Unlock()

	module := "ports"
	if t.module != "" {
		module = t.module
	}
	samples := t.cursor
	cursor := t.cursor
	if cursor > t.maxSamples {
		samples = t.maxSamples
	}
	v := &VCDDetail{
		module:   module,
		fullMask: t.fullMask,
		app:      t.app,
		samples:  samples,
		working:  make([]Sample, samples),
		start:    (cursor + t.maxSamples - samples) % t.maxSamples,
	}
	if samples != cursor {
		copy(v.working[:samples-v.start], t.samples[v.start:])
		copy(v.working[samples-v.start:], t.samples[:v.start])
	} else {
		copy(v.working[:samples], t.samples[:samples])
	}
	j := index
	for i, bit := 0, uint64(1); bit != 0 && bit < v.fullMask; i, bit = i+1, bit<<1 {
		if v.fullMask&bit != 0 {
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
			v.sigs = append(v.sigs, sig)
		}
	}
	return v, j
}

// Datum is used to manage the construction of a VCD signal trace dump.
type Datum struct {
	stamp uint
	line  string
}

func mergeVCD(earliest time.Time, tScale time.Duration, details []*VCDDetail) <-chan Datum {
	ch := make(chan Datum)
	go func() {
		defer close(ch)
		k := len(details)
		switch k {
		case 0:
			return
		case 1:
			var lastVal, lastMask uint64
			v := details[0]
			for i := uint(0); i < v.samples; i++ {
				s := v.working[i]
				delta := lastVal ^ s.Value
				dMask := lastMask ^ s.Mask
				if anyDelta := dMask | delta; i == 0 || anyDelta != 0 {
					// Something has changed, so we need to include it in the dump file.
					stamp := uint(s.When.Sub(earliest) / tScale)
					for _, sig := range v.sigs {
						if sig.mask&s.Mask == 0 {
							if i == 0 || sig.mask&dMask != 0 {
								ch <- Datum{
									stamp: stamp,
									line:  fmt.Sprint("x", sig.ch),
								}
							}
						} else {
							if i == 0 || sig.mask&anyDelta != 0 {
								v := 0
								if sig.mask&s.Value != 0 {
									v = 1
								}
								ch <- Datum{
									stamp: stamp,
									line:  fmt.Sprintf("%d%s", v, sig.ch),
								}
							}
						}
					}
				}
				lastVal = s.Value
				lastMask = s.Mask
			}
		default:
			j := (k + 1) / 2
			a := mergeVCD(earliest, tScale, details[:j])
			b := mergeVCD(earliest, tScale, details[j:])
			var lastA, lastB Datum
			var okA, okB bool
			for a != nil || b != nil {
				if a != nil && !okA {
					lastA, okA = <-a
					if !okA {
						a = nil
					}
				}
				if b != nil && !okB {
					lastB, okB = <-b
					if !okB {
						b = nil
					}
				}
				if okA && okB {
					if lastA.stamp <= lastB.stamp {
						ch <- lastA
						okA = false
					} else {
						ch <- lastB
						okB = false
					}
				} else if okA {
					ch <- lastA
					okA = false
				} else if okB {
					ch <- lastB
					okB = false
				}
			}
		}
	}()
	return ch
}

// ExportVCD generates a single VCD dump file from a set of concurrent
// trace recordings. The argument dumper names the collection of
// traces and tScale indicates what a count of 1 means in the counter
// output.
func ExportVCD(dumper string, tScale time.Duration, traces ...*Trace) (<-chan string, error) {
	var details []*VCDDetail
	j := 0
	var earliest time.Time
	for i, t := range traces {
		v, k := t.cacheVCDDetail(j)
		if k == j {
			log.Printf("skipping trace %d: no signals recorded", i)
			continue
		}
		details = append(details, v)
		if then := v.working[0].When; j == 0 || then.Before(earliest) {
			earliest = then
		}
		j = k
	}
	if j == 0 {
		return nil, ErrNoTraceData
	}
	ch := make(chan string)
	go func() {
		defer close(ch)

		vcdSection(ch, "date", earliest.Format(time.ANSIC), false)
		vcdSection(ch, "version", dumper, false)
		vcdSection(ch, "timescale", fmt.Sprintf("%v", tScale), false)

		for _, v := range details {
			vcdSection(ch, "scope", fmt.Sprintf("module %s", v.app), true)
			vcdSection(ch, "scope", fmt.Sprintf("module %s", v.module), true)
			for _, sig := range v.sigs {
				vcdSection(ch, "var", fmt.Sprintf("wire 1 %s %s", sig.ch, sig.lab), true)
			}
			ch <- "$upscope $end"
			ch <- "$upscope $end"
		}

		ch <- "$enddefinitions $end"

		var stamp uint
		started := false
		for datum := range mergeVCD(earliest, tScale, details) {
			if !started || datum.stamp != stamp {
				stamp = datum.stamp
				ch <- fmt.Sprint("#", stamp)
				if !started {
					ch <- "$dumpvars"
					started = true
				}
			}
			ch <- datum.line
		}
	}()
	return ch, nil
}

// VCD generates a Value Change Dump from the trace recorded so far.
// The function starts by making a snapshot of the current trace.
func (t *Trace) VCD(tScale time.Duration) (io.Reader, error) {
	if t == nil {
		return nil, ErrNoTraceData
	}

	ch, err := ExportVCD("iotracer", tScale, t)
	if err != nil {
		return nil, err
	}

	w := &bytes.Buffer{}
	for line := range ch {
		fmt.Fprintln(w, line)
	}

	return w, nil
}
