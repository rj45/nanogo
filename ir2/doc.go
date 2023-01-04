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
package ir2
