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
sample.ports.sig3 @ 2023-04-16 18:54:17.261259046 = true
sample.ports.sig3 @ 2023-04-16 18:54:17.261269046 = false
sample.ports.sig3 @ 2023-04-16 18:54:17.261285446 = true
sample.ports.sig3 @ 2023-04-16 18:54:17.261308246 = false
sample.ports.sig3 @ 2023-04-16 18:54:17.261312746 = true
sample.ports.sig3 @ 2023-04-16 18:54:17.261313746 = false
sample.ports.sig3 @ 2023-04-16 18:54:17.261314846 = true
sample.ports.sig3 @ 2023-04-16 18:54:17.261316046 = false
sample.ports.sig3 @ 2023-04-16 18:54:17.261317346 = true
sample.ports.sig3 @ 2023-04-16 18:54:17.261318746 = false
sample.ports.sig3 @ 2023-04-16 18:54:17.261320246 = true
sample.ports.sig3 @ 2023-04-16 18:54:17.261321846 = false
sample.ports.sig3 @ 2023-04-16 18:54:17.261340746 = true
sample.ports.sig3 @ 2023-04-16 18:54:17.261343346 = false
sample.ports.sig3 @ 2023-04-16 18:54:17.261346046 = true
sample.ports.sig3 @ 2023-04-16 18:54:17.261348846 = false
sample.ports.sig3 @ 2023-04-16 18:54:17.261351746 = true
sample.ports.sig3 @ 2023-04-16 18:54:17.261354746 = false
sample.ports.sig3 @ 2023-04-16 18:54:17.261357846 = true
sample.ports.sig3 @ 2023-04-16 18:54:17.261361046 = false
```

The timestamps and `true` and `false` values are tracking trace events
where `sample.ports.sig3` changes. In addition to this, the above
command generates a `dump.vcd` file.

NOTE: the dates in the above output refer to the actual timestamps at
the time `./sample` was running. VCD dump contains enough precision to
reconstruct these timestamps to the precision used to generate the VCD
dump. Specifically:
```
$ twave --file dump.vcd | head -30
[] : [$version top $end]
               sample.ports.sig0-+
               sample.ports.sig1-|-+
               sample.ports.sig2-|-|-+
               sample.ports.sig3-|-|-|-+
               sample.ports.sig4-|-|-|-|-+
               sample.ports.sig5-|-|-|-|-|-+
                other.ports.sig0-|-|-|-|-|-|-+
           other.ports.octo[0:2]-|-|-|-|-|-|-|---+
                other.ports.sig4-|-|-|-|-|-|-|---|-+
                other.ports.sig5-|-|-|-|-|-|-|---|-|-+
                                 | | | | | | |   | | |
2023-04-16 18:54:17.261255446000 x x x x x x x xxx x x
2023-04-16 18:54:17.261255546000 1 x x x x x x xxx x x
2023-04-16 18:54:17.261255746000 0 1 x x x x x xxx x x
2023-04-16 18:54:17.261256046000 1 1 x x x x x xxx x x
2023-04-16 18:54:17.261256446000 0 0 1 x x x x xxx x x
2023-04-16 18:54:17.261256946000 1 0 1 x x x x xxx x x
2023-04-16 18:54:17.261257546000 0 1 1 x x x x xxx x x
2023-04-16 18:54:17.261258246000 1 1 1 x x x x xxx x x
2023-04-16 18:54:17.261259046000 0 0 0 1 x x x xxx x x
2023-04-16 18:54:17.261259946000 1 0 0 1 x x x xxx x x
2023-04-16 18:54:17.261260946000 0 1 0 1 x x x xxx x x
2023-04-16 18:54:17.261262046000 1 1 0 1 x x x xxx x x
2023-04-16 18:54:17.261263246000 0 0 1 1 x x x xxx x x
2023-04-16 18:54:17.261264546000 1 0 1 1 x x x xxx x x
2023-04-16 18:54:17.261265946000 0 1 1 1 x x x xxx x x
2023-04-16 18:54:17.261266046000 0 1 1 1 x x 1 001 x x
2023-04-16 18:54:17.261266446000 0 1 1 1 x x 0 010 x x
2023-04-16 18:54:17.261266946000 0 1 1 1 x x 1 010 x x
```
See, that in the `dump.vcd` file (displayed with
[`twave`](https://github.com/tinkerator/twave)) the
`sample.ports.sig3` signal transitions at
`2023-04-16 18:54:17.261259046000` and in the original `./sample` run,
it is recorded at `2023-04-16 18:54:17.261259046`.

You can see how this trace looks using
[twave](https://github.com/tinkerator/twave), as shown above, or with
a more graphical experience with
[GTKWave](https://gtkwave.sourceforge.net/). For the latter, `gtkwave
dump.vcd` and selecting all of the signals displays as follows:

![GTKWave rendering of this `dump.vcd` file.](screenshot.png)

## License info

The `iotracer` package is distributed with the same BSD 3-clause license
as that used by [golang](https://golang.org/LICENSE) itself.

## Reporting bugs and feature requests

The package `iotracer` has been developed purely out of self-interest and
a curiosity for debugging physical IO projects, primarily on the
Raspberry Pi. Should you find a bug or want to suggest a feature
addition, please use the [bug
tracker](https://github.com/tinkerator/iotracer/issues).
