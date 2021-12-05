# NanoGo

`NanoGo` is a Go (golang) compiler for homebrew / hobby CPUs.

## Status

Note: this is a work in progress! Currently the following should work:

- built in `print()` and `println()`
- word-sized operations (`int` and `uint`)
- string literals and iterating over strings
- memory mapped I/O using `unsafe`
- extern funcs with assembly snippets (useful if you have I/O instructions)

Also, only [rj32](https://github.com/rj45/rj32) and [A32](https://github.com/Artentus/a32emu) are supported, but if you would like assistance adding your CPU, open an issue.

## Installing

Install [Go](https://golang.org/) for your system, then:

```sh
$ go get -u github.com/rj45/nanogo
```

You will also want to install [customasm](https://github.com/hlorenzi/customasm).

You may also want to install [emurj](https://github.com/rj45/rj32/emurj) if you want to be able to run programs in an `rj32` emulator.

## What is it?

This compiler will take a Go package, read in all the packages it depends on in the usual way Go programs work, and compile all the code down into assembler in a style that is compatible with [customasm](https://github.com/hlorenzi/customasm).

As of this writing, customasm does not support linking, so a single large assembly file is produced. A "CPU Def" file can be included which configures the assembly language, as well as the memory layout with `#bank`s.

## Why Go?

C is great, but the language is not the easiest to parse, and while there's many great projects like [LCC](https://github.com/drh/lcc), they are not the easiest to work on and modify for a homebrew CPU.

Go is a very simple language, and is very fast to learn, yet powerful enough to be very productive.

As well, Go has a parser and type checker right in the the standard library. Further, there's an excellent [SSA library](https://golang.org/x/tools/go/ssa) that does most of the work of the frontend of a compiler. So all that's really required is a simple backend, which is what NanoGo is.

## Limitations

While all of Go is parsed, obviously there's no garbage collector, and no goroutines.

It took a team many years to build the Go compiler, so as this is a hobby project primarily written by one person, there's a lot left completely unimplemented, and error handling could be a lot better.

Defer is ignored, though it could be implemented in the future. Recovering from panics will not be implemented. Runtime type reflection is not yet implemented. Maps are not yet implemented. Interfaces are similarly not there, nor slices.

`int`s, `uint`s, `byte`s, `rune`s and pointers are 16-bits for rj32. But non-standard sizes can violate some assumptions in the standard library, so anything relying on those assumptions will have bugs.

Almost none of the standard library is supported. You can try it and see if it will work, but some fundamental assumptions are violated, as well as many features relied are missing. This compiler is meant to help you write your own standard library, kernel, OS and other software for your own homebrew CPU, so you could see the lack of a standard library as a feature.

## Design

NanoGo uses the [golang.org/x/tools/go/ssa](https://golang.org/x/tools/go/ssa) package to build and parse Go into [SSA form](https://en.wikipedia.org/wiki/Static_single_assignment_form).

This IR is then [converted](./parser) into a different [IR](./ir) which is [easier to work with](https://www.hpl.hp.com/techreports/Compaq-DEC/WRL-2000-2.html) and [transform](./xform).

After transformation there is a [SSA based tree register allocator](./regalloc) which uses the properties of SSA to [allocate registers in linear time](https://compilers.cs.uni-saarland.de/projects/ssara/).

After a final transformation pass, then finally [assembly code is generated](./codegen) and output.

## Retargetting / Porting to Your CPU

You can start by copying an [existing arch](./arch) and modifying it for your instruction set.

I will be writing more about this process soon. In the meantime, if you would like some assistance, open an issue.

## License

Copyright (c) 2021 rj45 ([github.com/rj45](https://github.com/rj45))

Licensed under the MIT License (MIT), see [LICENSE](./LICENSE).

Some parts Copyrighted by the Go Authors, under a BSD like license, see [Go's LICENSE](https://github.com/golang/go/blob/666fc173c02ff3004ac9ef867aa4eec7e243dde3/LICENSE).

Some parts Copyrighted by the TinyGo Authors, under a BSD like license, see [TinyGo's LICENSE](./LICENSE.tinygo).
