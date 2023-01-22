# iotracer - a package for tracing digital signals.

## Overview

When you have values that change over time, it is useful to be able to
track them. This is especially useful for digital signals, such as
GPIO values. Debugging problems with these signals firing when you
don't expect them to, or not when you do is much easier if you can
trace them as they change.

This package provides a circular buffer implementation of a tracing
facility. It also supports dumping the content of the trace buffer in
VCD [Value Change
Dump](https://en.wikipedia.org/wiki/Value_change_dump) format.

Automated package documentation for this Go package should be
available from [![Go
Reference](https://pkg.go.dev/badge/zappem.net/pub/io/iotracer.svg)](https://pkg.go.dev/zappem.net/pub/io/iotracer).

## Getting started

You can build an example and run it as follows:
```
$ go build examples/sample.go
$ ./sample > dump.vcd
ports.sig3 @ 2023-01-20 21:55:37.174626915 = true
ports.sig3 @ 2023-01-20 21:55:37.174636915 = false
ports.sig3 @ 2023-01-20 21:55:37.174653315 = true
ports.sig3 @ 2023-01-20 21:55:37.174676115 = false
ports.sig3 @ 2023-01-20 21:55:37.174680615 = true
ports.sig3 @ 2023-01-20 21:55:37.174681615 = false
ports.sig3 @ 2023-01-20 21:55:37.174682715 = true
ports.sig3 @ 2023-01-20 21:55:37.174683915 = false
ports.sig3 @ 2023-01-20 21:55:37.174685215 = true
ports.sig3 @ 2023-01-20 21:55:37.174686615 = false
ports.sig3 @ 2023-01-20 21:55:37.174688115 = true
ports.sig3 @ 2023-01-20 21:55:37.174689715 = false
ports.sig3 @ 2023-01-20 21:55:37.174708615 = true
ports.sig3 @ 2023-01-20 21:55:37.174711215 = false
ports.sig3 @ 2023-01-20 21:55:37.174713915 = true
ports.sig3 @ 2023-01-20 21:55:37.174716715 = false
ports.sig3 @ 2023-01-20 21:55:37.174719615 = true
ports.sig3 @ 2023-01-20 21:55:37.174722615 = false
ports.sig3 @ 2023-01-20 21:55:37.174725715 = true
ports.sig3 @ 2023-01-20 21:55:37.174728915 = false
```

The timestamps and `true` and `false` values are tracking trace events
where `sig3` changes. In addition to this, the above command generates
a `dump.vcd` file.

You can see how this trace looks using
[twave](https://github.com/tinkerator/twave), or as a more friendly
experience with [GTKWave](https://gtkwave.sourceforge.net/). For the
latter, `gtkwave dump.vcd` and selecting all of the signals displays
as follows:

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
