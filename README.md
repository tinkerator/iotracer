# iotracer - a package for tracing digital signals.

## Overview

When you have values that change over time, it is useful to be able to
track them. This is especially useful for digital signals, such as
GPIO values. Debugging problems with these signals firing when you
don't expect them to, or not when you do is much easier if you can
trace them as they change.

This package provides a circular buffer implementation of a tracing
facility. It also supports dumping the content of the trace buffer in
[Value Change Dump](https://en.wikipedia.org/wiki/Value_change_dump)
(VCD) format.

Automated package documentation for this Go package should be
available from [![Go
Reference](https://pkg.go.dev/badge/zappem.net/pub/io/iotracer.svg)](https://pkg.go.dev/zappem.net/pub/io/iotracer).

## Getting started

You can build an example and run it as follows:
```
$ go build examples/sample.go
$ ./sample > dump.vcd
sample.ports.sig3 @ 2023-01-28 18:48:19.711618752 = true
sample.ports.sig3 @ 2023-01-28 18:48:19.711628752 = false
sample.ports.sig3 @ 2023-01-28 18:48:19.711645152 = true
sample.ports.sig3 @ 2023-01-28 18:48:19.711667952 = false
sample.ports.sig3 @ 2023-01-28 18:48:19.711672452 = true
sample.ports.sig3 @ 2023-01-28 18:48:19.711673452 = false
sample.ports.sig3 @ 2023-01-28 18:48:19.711674552 = true
sample.ports.sig3 @ 2023-01-28 18:48:19.711675752 = false
sample.ports.sig3 @ 2023-01-28 18:48:19.711677052 = true
sample.ports.sig3 @ 2023-01-28 18:48:19.711678452 = false
sample.ports.sig3 @ 2023-01-28 18:48:19.711679952 = true
sample.ports.sig3 @ 2023-01-28 18:48:19.711681552 = false
sample.ports.sig3 @ 2023-01-28 18:48:19.711700452 = true
sample.ports.sig3 @ 2023-01-28 18:48:19.711703052 = false
sample.ports.sig3 @ 2023-01-28 18:48:19.711705752 = true
sample.ports.sig3 @ 2023-01-28 18:48:19.711708552 = false
sample.ports.sig3 @ 2023-01-28 18:48:19.711711452 = true
sample.ports.sig3 @ 2023-01-28 18:48:19.711714452 = false
sample.ports.sig3 @ 2023-01-28 18:48:19.711717552 = true
sample.ports.sig3 @ 2023-01-28 18:48:19.711720752 = false
```

The timestamps and `true` and `false` values are tracking trace events
where `sample.ports.sig3` changes. In addition to this, the above
command generates a `dump.vcd` file.

NOTE: the dates in the above output refer to the actual timestamps at
the time `./sample` was running. However, the VCD dump does not
contain enough precision for the initial timestamp to reproduce these
timestamps exactly. Only the relative timestamp markers between signal
transitions are preserved. Specifically:
```
$ twave --file dump.vcd | head -25
[] : [$version top $end]
               sample.ports.sig0-+
               sample.ports.sig1-|-+
               sample.ports.sig2-|-|-+
               sample.ports.sig3-|-|-|-+
               sample.ports.sig4-|-|-|-|-+
               sample.ports.sig5-|-|-|-|-|-+
                other.ports.sig0-|-|-|-|-|-|-+
                other.ports.sig1-|-|-|-|-|-|-|-+
                other.ports.sig2-|-|-|-|-|-|-|-|-+
                other.ports.sig3-|-|-|-|-|-|-|-|-|-+
                other.ports.sig4-|-|-|-|-|-|-|-|-|-|-+
                other.ports.sig5-|-|-|-|-|-|-|-|-|-|-|-+
                                 | | | | | | | | | | | |
2023-01-28 18:48:19.000000000000 x x x x x x      
2023-01-28 18:48:19.000000100000 1 x x x x x      
2023-01-28 18:48:19.000000300000 0 1 x x x x      
2023-01-28 18:48:19.000000600000 1 1 x x x x      
2023-01-28 18:48:19.000001000000 0 0 1 x x x      
2023-01-28 18:48:19.000001500000 1 0 1 x x x      
2023-01-28 18:48:19.000002100000 0 1 1 x x x      
2023-01-28 18:48:19.000002800000 1 1 1 x x x      
2023-01-28 18:48:19.000003600000 0 0 0 1 x x      
2023-01-28 18:48:19.000004500000 1 0 0 1 x x      
2023-01-28 18:48:19.000005500000 0 1 0 1 x x      
```
See, that in the `dump.vcd` file (displayed with
[`twave`](https://github.com/tinkerator/twave)) the
`sample.ports.sig3` signal transitions at
`2023-01-28 18:48:19.000003600000` but in the original `./sample` run,
it is recorded at `2023-01-28 18:48:19.711618752`. This is not a bug,
but an artifact of the VCD file format.

You can see how this trace looks using
[twave](https://github.com/tinkerator/twave), as shown above, or with
a more graphical experience with
[GTKWave](https://gtkwave.sourceforge.net/). For the latter, `gtkwave
dump.vcd` and selecting all of the signals displays as follows:

![GTKWave rendering of this `dump.vcd` file.](screenshot.png)

## TODOs

- Support more than one bank of VCD dumping at a time.
- Some more sophisticated output labeling perhaps by grouping of
  labels where it makes sense to capture numerical values that are
  longer than a single bit.

## License info

The `iotracer` package is distributed with the same BSD 3-clause license
as that used by [golang](https://golang.org/LICENSE) itself.

## Reporting bugs and feature requests

The package `iotracer` has been developed purely out of self-interest and
a curiosity for debugging physical IO projects, primarily on the
Raspberry Pi. Should you find a bug or want to suggest a feature
addition, please use the [bug
tracker](https://github.com/tinkerator/iotracer/issues).
