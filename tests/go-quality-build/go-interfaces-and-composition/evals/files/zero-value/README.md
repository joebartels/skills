# Bag

Bag accumulates event-name counts inside a single worker. The zero value is ready for use, including when embedded in another struct. Count of an unseen key is zero. All strings, including empty, are valid keys. Callers do not share one Bag across goroutines. Its map is private, and callers never allocate that map.

The requested limit is a nonnegative count of distinct keys; zero means unlimited. A positive limit remains in force after clearing. Configuration is chosen before first use; changing it on a populated Bag is outside this change. Counts in this application fit in int. This is internal code targeting Go 1.22; callers can be updated in the same change.
