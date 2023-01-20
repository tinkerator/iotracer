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

## Getting started

You can build an example and run it as follows:
```
$ go build examples/sample.go
$ ./sample > dump.vcd
```

You can see how this trace looks using
[twave](https://github.com/tinkerator/twave), or as a more friendly
experience with [GTKWave](https://gtkwave.sourceforge.net/). For the
latter, `gtkwave dump.vcd` and selecting all of the signals displays
as follows:

![GTKWave rendering of this `dump.vcd` file.](screenshot.png)

## TODOs

- more sophisticated output labeling perhaps even grouping of labels
  where it makes sense to capture numerical values that are longer
  than a single bit.

## License info

The `iotracer` package is distributed with the same BSD 3-clause license
as that used by [golang](https://golang.org/LICENSE) itself.

## Reporting bugs and feature requests

The package `iotracer` has been developed purely out of self-interest and
a curiosity for debugging physical IO projects, primarily on the
Raspberry Pi. Should you find a bug or want to suggest a feature
addition, please use the [bug
tracker](https://github.com/tinkerator/iotracer/issues).
