# Gismo Example: Custom C-Like Language Compiler

This directory contains a complete example of how to use Gismo as a **Meta-Compiler Framework**. Instead of writing a parser and lexer from scratch, we use Gismo's "Interpiler" architecture to define a statically typed, C-like language that compiles down to a register-based Intermediate Representation (IR).

## File Structure

* **`main.gsm`**: The user-space code written in our custom language. This is the "source" file that gets compiled.
* **`toolchain/language.gsm`**: The core language definition. It defines how function declarations, variable assignments, and arithmetic operators are transformed into target IR code.
* **`toolchain/types.gsm`**: Defines the meta-structures for the compiler (Functions, Variables, Values) using Gismo's struct system.
* **`toolchain/before.gsm`**: The standard library providing vector operations and basic primitives required by the toolchain.

## Usage

To compile the `main.gsm` file and generate the SSA Intermediate Representation, use the following command:

```bash
compiler main.gsm -o out.ssa
```

This will execute the meta-definitions in the toolchain, process your code, and save the generated IR to `out.ssa`.

## How It Works

Gismo executes the `toolchain` files first, which set up the rules for syntax and code generation. When the interpreter encounters the syntax in `main.gsm`, it triggers these rules to emit code immediately (Compile-Time Execution).

### 1. Defining Syntax (The "Interpiler" Logic)

In `toolchain/language.gsm`, we use Gismo's `::=` operator to intercept standard syntax and generate IR instructions.

**Example: Defining Addition**
Instead of calculating a result, the `+` operator is overloaded to write an `add` instruction to the output stream.

```gismo
// When adding two runtime values (VALUE_I32), generate a new temporary register
VALUE_I32 + VALUE_I32 ::= {
    value_a ::= $UNTYPE($1)
    value_b ::= $UNTYPE($2)
    addr ::= $CAT("%t", $IOTA()) // Create unique register %t0, %t1...

    // Emit the target assembly/IR
    $WRITE("\t")
    $WRITE(addr)
    $WRITE(" =w add ")
    $WRITE(value_a.addr)
    $WRITE(", ")
    $WRITE(value_b.addr)
    $WRITE("\n")

    // Return a typed object representing the result register
    $TYPEDEF($Value(i32, addr), VALUE_I32)
}
```

### 2. User Code (`main.gsm`)

The user writes code that looks like a high-level language. Because of the rules defined above, this valid Gismo syntax acts as a DSL.

```gismo
add(a: i32, b: i32): i32 {
    sum2(a, b) // Calls another function
}

main(): i32 {
    a <- 20    // Variable assignment (defined via 'symbol <- int')
    b <- 30
    add(a, b)
}
```

### 3. Generated Output (Target IR)

When `main.gsm` is executed by Gismo, it produces the following Intermediate Representation to stdout or the specified output file (`out.ssa`):

```asm
function add(i32 %a, i32 %b) {
@start
    %t0 =w call $sum2(%a, %b)
    ret %t0
}

export function $main() {
@start
    %a =w copy 20
    %b =w copy 30
    %t1 =w call $add(%a, %b)
    ret %t1
}
```

## Key Gismo Features Demonstrated

* **Compile-Time Dispatch**: Using `symbol(...)` patterns to handle function declarations dynamically.
* **Code Emission (`$WRITE`)**: Explicit control over the output artifact.
* **Meta-Typing**: Using `$TYPEDEF` to create a distinction between compile-time integers (`int`) and runtime target values (`VALUE_I32`).
* **Homoiconicity**: The compiler is written in the same language syntax as the target DSL.
