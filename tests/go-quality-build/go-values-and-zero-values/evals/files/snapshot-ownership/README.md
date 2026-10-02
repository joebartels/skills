# editgraph
Go 1.22 library for editable object graphs. Nodes may share link targets or have
cycles; callers own nodes and may edit them. Graph is a lightweight value handle.
RootView deliberately borrows the original root: callers use it for in-place
editing. Keep that behavior and Graph's existing value-method/interface usability.
Add Graph.Fork() Graph to make an independently editable draft of the reachable
graph at capture time. Mutations to names, map entries, byte payloads, or links on
either graph must not affect the other. A fork retains the original graph layout:
links to the same node still share one corresponding node in the fork, including
cycles. Preserve nil versus non-nil empty collections and a usable zero Graph.
No concurrency guarantee or forced constructor is required. Document the
borrowed view and owned fork; add meaningful tests without redesigning the API.
