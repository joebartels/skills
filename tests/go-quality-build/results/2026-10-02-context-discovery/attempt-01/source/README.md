# Sequential operation

Update Process and its tests to meet this contract. Keep its signature and Go 1.22 support; do not introduce concurrency or dependencies.

Process accepts a nonnil context, integer jobs and a nonnil cooperative callback. Invoke jobs sequentially. A nil callback result accepts that job; return the exact accepted count on every exit. Already-canceled calls invoke no callback and return cancellation classification/cause. Otherwise empty input succeeds with zero accepted.

Check cancellation before starting another callback. On a callback error, decide the result immediately after that error is returned: retain the independent callback error, and if cancellation has already been observed there, also retain its standard classification and inspectable custom cause. Causes may be any legal error value, including non-comparable concrete values. No equality comparison may panic. Preserve independent failure rather than substituting a cancellation-only result.

A nil last-callback result completes the operation successfully even when cancellation coincides with that accepted completion. Cancellation after completion must not invalidate it. Between-job cancellation stops later calls with the accepted prefix and classification/cause. Cancellation requests stopping; the callback remains responsible for returning cooperatively. Caller owns its context and callback resources.
