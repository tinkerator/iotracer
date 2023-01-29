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
sample.ports.sig3 @ 2023-01-29 13:39:00.832526747 = true
sample.ports.sig3 @ 2023-01-29 13:39:00.832536747 = false
sample.ports.sig3 @ 2023-01-29 13:39:00.832553147 = true
sample.ports.sig3 @ 2023-01-29 13:39:00.832575947 = false
sample.ports.sig3 @ 2023-01-29 13:39:00.832580447 = true
sample.ports.sig3 @ 2023-01-29 13:39:00.832581447 = false
sample.ports.sig3 @ 2023-01-29 13:39:00.832582547 = true
sample.ports.sig3 @ 2023-01-29 13:39:00.832583747 = false
sample.ports.sig3 @ 2023-01-29 13:39:00.832585047 = true
sample.ports.sig3 @ 2023-01-29 13:39:00.832586447 = false
sample.ports.sig3 @ 2023-01-29 13:39:00.832587947 = true
sample.ports.sig3 @ 2023-01-29 13:39:00.832589547 = false
sample.ports.sig3 @ 2023-01-29 13:39:00.832608447 = true
sample.ports.sig3 @ 2023-01-29 13:39:00.832611047 = false
sample.ports.sig3 @ 2023-01-29 13:39:00.832613747 = true
sample.ports.sig3 @ 2023-01-29 13:39:00.832616547 = false
sample.ports.sig3 @ 2023-01-29 13:39:00.832619447 = true
sample.ports.sig3 @ 2023-01-29 13:39:00.832622447 = false
sample.ports.sig3 @ 2023-01-29 13:39:00.832625547 = true
sample.ports.sig3 @ 2023-01-29 13:39:00.832628747 = false
```

The timestamps and `true` and `false` values are tracking trace events
where `sample.ports.sig3` changes. In addition to this, the above
command generates a `dump.vcd` file.

NOTE: the dates in the above output refer to the actual timestamps at
the time `./sample` was running. VCD dump contains enough precision to
reconstruct these timestamps to the precision used to generate the VCD
dump. Specifically:
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
2023-01-29 13:39:00.832523147000 x x x x x x      
2023-01-29 13:39:00.832523247000 1 x x x x x      
2023-01-29 13:39:00.832523447000 0 1 x x x x      
2023-01-29 13:39:00.832523747000 1 1 x x x x      
2023-01-29 13:39:00.832524147000 0 0 1 x x x      
2023-01-29 13:39:00.832524647000 1 0 1 x x x      
2023-01-29 13:39:00.832525247000 0 1 1 x x x      
2023-01-29 13:39:00.832525947000 1 1 1 x x x      
2023-01-29 13:39:00.832526747000 0 0 0 1 x x      
2023-01-29 13:39:00.832527647000 1 0 0 1 x x      
2023-01-29 13:39:00.832528647000 0 1 0 1 x x
```
See, that in the `dump.vcd` file (displayed with
[`twave`](https://github.com/tinkerator/twave)) the
`sample.ports.sig3` signal transitions at
`2023-01-29 13:39:00.832526747000` and in the original `./sample` run,
it is recorded at `2023-01-29 13:39:00.832526747`.

You can see how this trace looks using
[twave](https://github.com/tinkerator/twave), as shown above, or with
a more graphical experience with
[GTKWave](https://gtkwave.sourceforge.net/). For the latter, `gtkwave
dump.vcd` and selecting all of the signals displays as follows:

![GTKWave rendering of this `dump.vcd` file.](screenshot.png)

## TODOs

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
