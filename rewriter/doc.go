// This is a rewrite engine language inspired by
// cranelift's isle language:
// https://cfallin.org/blog/2023/01/20/cranelift-isle/
//
// The core idea is instead of matching node types directly,
// a match is done using function calls. This is more flexible
// because it can be adapted to work in more situations.
//
// Unlike ISLE, though, it's not strongly typed, and doesn't have
// extra language features for defining types. Because, well,
// this is a hobby project.
//
// A function will be generated that takes a matcher "iterator" and
// a builder. Rule patterns will be translated into matcher calls,
// and it's expected each matcher call will return matchers for sub-trees,
// as well as a final bool result indicating if the match was successful.
//
// Two special functions on the matcher `Remove` and `Replace` will be called
// to remove the current node, or replace it respectively.
// Replace takes a reference to a matcher for the node to replace it.
//
// The builder is used to build up a new sub-tree given matchers for parts of the
// sub-tree.
package rewriter
