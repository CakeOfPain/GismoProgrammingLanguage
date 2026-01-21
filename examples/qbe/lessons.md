# Lesson that I have learn by programming in my language

- Gismo is very unsafe at the moment
    - You can use these CompileTime Inbuilts that are not compatible with the whole type polymorphism idea
    - When using `$TYPEDEF(..., TYPENAME, ..., OTHER_TYPES)` the typename can be choosen as you wish, without making sure, it will be the same across the other `$TYPEDEF`s.

- It would be cool to let gismo decide what it will load as a main file by passing the environment argument to the interpreter
- The ability to load the ast of a file not executing it

- there is no hierachy of types within the inbuilts (at least `any` would be helpful!)
- Once a value is wrapped inside a type value, you have to unwrap it all the time, which is very exhausting
- No more workarounds using `$EVAL`... Only clean syntax
- Inbuilt functions should have a clear name convention, that is checked by the compiler!
- I need a linter (camelCase & snake_case Problem)
- It should be hard to destory the compiler by `$WRITE` instructions or compile time exceptions
- The whole Lisp `$CONS` has proven to be exhausting for datatype making
- The compile time code is very unreadable

