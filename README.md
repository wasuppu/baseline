# baseline 

This is a Go implementation of [Compiling to Assembly from Scratch](https://keleshev.com/compiling-to-assembly-from-scratch/) Part I: Baseline Compiler.

I haven't verified the correctness or executability of the generated ARM assembly. 
What's interesting about this book is that its parsing section applays combinator parsing, 
a concept I've encountered in Haskell but never used in imperative languages before. 
While the idea is conceptually clear and simple, I actually had to try multiple times before getting it to work correctly in Go.
