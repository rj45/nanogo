// Package ir2 contains an Intermediate Representation (IR) in
// Static Single Assignment (SSA) form.
//
//   - Each Program contains a list of Packages.
//   - Packages are a list of Globals, TypeDefs and Funcs.
//   - Each Func is a list of Blocks.
//   - Blocks have a list of Instrs.
//   - Instrs Def (define) Values, and have Values as Args.
//   - Values can be constants, types, temps, registers, memory locations, etc.
//
// Note: Unlike other SSA representations, this representation
// separates the concept of instructions from the concept of
// values. This allows an instruction to define multiple values.
// This is handy to avoid needing tuples and unpacking tuples to
// handle instructions (like function calls) that return multiple
// values.
//
// This IR is structured with ideas from Data Oriented Programming
// and Entity Component Systems type thinking to try to structure
// data to be close together in cache and avoid cache misses if
// possible.
//
// As such there is an ID system that acts like a index into an
// array. These IDs are local to the Func in which they live. The
// different kinds of ID-able things are allocated in slabs of a
// fixed size in order to avoid invalidating pointers. Memory
// currently is not freed (but could be with a generation counter in
// the ID.)
//
// The IR is designed to be serialized into a human readable text file
// and parsed back into IR to aid in creating tests.
package ir2
