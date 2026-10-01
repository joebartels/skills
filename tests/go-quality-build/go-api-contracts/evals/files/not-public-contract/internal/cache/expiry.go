package cache

// expired reports whether the local cache entry's deadline has passed.
// A zero deadline means the entry never expires.
func expired(now, deadline int64) bool { return deadline != 0 && now > deadline }
