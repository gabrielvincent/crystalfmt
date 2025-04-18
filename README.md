# crystalfmt

A code formatter for the Crystal programming language written in Go

I created this project because when I first tried Crystal I thought there were
no code formatting tools for it. It turns out there is one baked in the
compiler.

This is far from supporting the complete set of Crystal features, but it works
most of the time. For the stuff that it doesn't support, it just writes that
bit without formatting.

I'm abandoning this now that I learned that I can simply do `crystal tool
format`.

But, hey, this was fun and I got find a good use for Go iterators!
