For this Go behavior fix, I selected and opened:

- `go-behavior-tests`, because the task changes behavior and asks for a focused regression test.
- `go-api-contracts`, because the inclusive upper-bound behavior may be part of a consumer-visible contract; the task description doesn’t say whether the predicate is exported.

I decline `go-test-isolation`: the described predicate is pure and the test has no shared state, external dependencies, or lifecycle concerns. The context, composition, and package-boundary skills are also irrelevant to the described change.

I stopped after skill selection. I did not inspect task code, edit files, run tests, browse externally, or delegate.