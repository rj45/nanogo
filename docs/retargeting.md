# Retargeting

You can start by copying an [existing arch](./arch) and modifying it for your instruction set. You may want to pick the one that's closest to your architecture to make this easier.

The architecture calls `var _ = arch.Register(cpuArch{})` to register with the arch package, but that won't trigger unless the package is imported somewhere. [compiler_test.go](../compiler/compiler_test.go) is one such place. [nanogo.go](../nanogo.go) is the other.

## Tagging and transforms

The transforms have a [tagging system](../xform/xform.go) in place for being able to turn them on/off for specific architectures. If a xform func has no tags, it is always active. Otherwise all of its tags must be present in the architecture's `XformTags()` list. Make sure not to break other architectures when adding new tags to xform functions.

Do note: Many bugs are caught in the register allocator's verifier, but that does not necessarily mean the allocator is broken. Often it can be a bad xform or an unimplemented feature. The verifier is just analyzing the value flow, so that is why it catches these sorts of issues.

## ABI

The ABI is Go specific. This compiler is meant to manage compiling the entire application for you, and is not meant to play nice with existing code or code generated from other compilers. That said, there is a way to refer to assembly code from Go, see the [runtime library](../src/runtime/) for an example. This could be used as a bridge to an existing code base.

## Debugging

Many parts of the compiler can emit more debugging info to greatly accelerate debugging. `-dump` can be used to dump a function to `ssa.html` where you can investigate how certain code was generated and in what stage. There's also `-dot` for generating control flow and value flow graphs. There's also various `-debug*` switches for logging details, especially of the register allocator.

Again, if the register allocator's verifier produces errors, it just means there's a problem with the flow of values in the program. Many things can cause this, including bad xforms and use of unimplemented features. The register allocator is currently fairly stable, so it's usually not a bug there, but register allocators are complex so it could still have bugs.

## Testing

You will also want to have a working emulator that will be able to exit with an error code when it encounters a `panic()`, ideally there should also be a way to write to stdout from the emulated program -- either by memory mapped IO (like rj32 does) or via in/out instructions (like a32 does).

You will want to add some assembly for outputting to the console. Extern funcs trigger a scan of the containing folder to check if there are .asm files tagged with the arch that might have assembly for those funcs. You can find examples in the [runtime library](../src/runtime/).

For automated testing, I have been using integration tests in the [testdata](../testdata/) folder, and adding them to [compiler_test.go](../compiler/compiler_test.go). The [coverall](../coverall.sh) script can be used to generate a coverage report. If you prefer to write unit tests rather than integration tests, that is welcome as well.

