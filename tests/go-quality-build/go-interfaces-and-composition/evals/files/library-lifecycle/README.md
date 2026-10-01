# refreshkit

An embedded library used by a desktop sync host and a server. Refresher.Refresh currently performs one synchronous refresh using a host callback. New validates that a callback exists. The host owns any network clients or files captured by the callback and releases them only after refresh work has returned. Callback implementations honor context cancellation. One-shot Refresh forwards that context and returns the callback error unchanged.

Requested periodic behavior: the host decides when each run begins and ends. Perform the first refresh immediately, then wait the supplied positive interval after each successful completion. No overlapping callbacks for a single run. Stop a run on its first callback error and report that error. Stop requests must reach an active callback. The host must be able to wait for completion before freeing resources, and stopping one refresher must not stop another. Each instance has at most one periodic run at a time; concurrent starts/restarts need not be supported. A nonpositive interval is a caller error. The periodic cancellation result must be documented.

This library is one component among several inside the host process. The executable coordinates process termination and resource teardown. Go 1.22 is supported. The API is pre-1.0; add the periodic API while retaining New and Refresh.
