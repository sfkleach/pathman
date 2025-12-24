# 0000 - Why Go, 2025-12-24

## Issue

Decide on the programming language to implement `pathman` in.

## Factors

- Fast compiler, suitable for a live demo of app development in a timeboxed window
- No runtime dependencies, ideally a self-contained binary distribution 
- Cross-platform support
- Fast startup time
- Strong standard library for file operations

## Options and Outcome

I considered both Rust and Go for this project. I regarded both as perfectly
appropriate implementation technologies. In the end I selected Go purely on
the basis of compilation speed and better fit for interactive demonstration of
developing code with AI assistance.

## Pros and Cons of Options

### Option 1: Rust

**Pros:**
- Memory safety
- Excellent performance
- No runtime dependencies - single binary
- Strong type system prevents many bugs at compile time
- Growing ecosystem with excellent tooling (cargo)
- Cross-platform support
- Great for systems programming

**Cons:**
- Slower compilation times (significant for interactive development demos)
- Steeper learning curve
- More verbose error handling
- Longer time to get a working prototype
- Borrow checker can slow down rapid prototyping
- More complex for a relatively simple utility

**Interesting:**
- Would be educational for demonstrating Rust development with AI
- Strong compile-time guarantees might catch more edge cases

### Option 2: Go (Selected)

**Pros:**
- Very fast compilation (crucial for live development demos)
- Simple, readable syntax - easy to demonstrate and explain
- No runtime dependencies - single static binary
- Excellent cross-platform support with easy cross-compilation
- Fast startup time (important for PATH generation on every shell start)
- Strong standard library for file operations, especially `os` and `filepath` packages
- Built-in testing framework
- Easy to onboard contributors (simpler language)
- Good tooling (`go fmt`, `go vet`, `go test`)
- Interfaces and composition work well for the modular design needed

**Cons:**
- Garbage collection adds runtime overhead and unpredictable pauses
- Less powerful type system than Rust
- Error handling can be verbose (`if err != nil`)
- Handling nil is tedious and error-prone

**Interesting:**
- Go's simplicity makes it ideal for demonstrating AI-assisted development
- The standard library covers our needs without external dependencies
- Fast iteration cycle matches the timeboxed development approach
- Good balance between safety and productivity

## Additional Notes

The decision came down to development velocity and demonstration suitability.
While Rust would provide stronger guarantees and potentially better performance,
Go's compilation speed and simplicity make it far more suitable for:

1. **Interactive development sessions**: Compilation in <1 second vs several seconds matters when demonstrating iterative development
2. **AI-assisted development**: Simpler syntax means clearer communication with AI tools and easier-to-explain code
3. **Time-boxed demonstrations**: Getting from concept to working prototype is faster

For pathman's use case (file operations, symlink management, PATH manipulation),
Go's performance is more than adequate. The fast startup time is actually
critical since `pathman path` is called on every shell initialization.

The single binary distribution was a hard requirement met by both languages, but
Go's built-in cross-compilation (`GOOS=windows GOARCH=amd64 go build`) is
particularly straightforward.

If time allowed, a demonstration of developing pathman in Rust would be a 
very worthwhile alternative.
