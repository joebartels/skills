# framecollect
Go 1.22 library. Source.Next returns a byte view that is valid only until the next
Next call. io.EOF ends normally; other errors preserve their cause and accepted
prefix. Collect already returns retained frames. Add CollectMatching(src Source,
keep func([]byte) bool) ([][]byte,error), preserving Collect's exact function type.
A nil predicate retains every frame. Invoke the predicate on the borrowed view
without an unnecessary pre-predicate copy; predicate callers may inspect but
must not mutate or retain that view. Each accepted returned frame must remain
stable after later Next calls and after the source buffer is reused. Nil frames
remain nil, non-nil empty frames remain non-nil. Source ownership is borrowed;
there is no Close operation or new lifecycle. Document and test the new behavior.
