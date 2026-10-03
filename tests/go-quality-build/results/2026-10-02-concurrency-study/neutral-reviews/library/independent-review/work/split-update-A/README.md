# Coherent totals

Keep the public Snapshot, Totals, Add and Snapshot interfaces. Go 1.22 minimum; standard library only. Totals' zero value is usable. A used Totals must not be copied. Add(delta) increments Count once and adds delta to Sum as one observable update. Concurrent Add/Snapshot calls are supported; every returned Snapshot is a coherent independent value, never a mix of updates. With delta 1, Count==Sum on every snapshot. Different instances have independent state. Keep a direct implementation; no new public API, framework or lock-free redesign.
