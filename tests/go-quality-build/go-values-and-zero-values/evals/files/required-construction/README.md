# signedcounter
Go 1.22 package. Counter requires a validated secret key via New; its zero value
is intentionally unsupported. A constructed counter owns its key and may be
shared by goroutines. The internal mutex must not be copied after first use.
Add Count() uint64 for concurrent-safe observation of the number of Sign calls.
Keep the constructor precondition and existing Sign signatures/behavior. Do not
make callers normalize keys or start using a zero Counter. Add tests and short
accurate docs. There is no requirement for a clone or alternate constructor.
