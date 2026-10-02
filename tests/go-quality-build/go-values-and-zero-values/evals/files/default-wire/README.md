# reportview
Go 1.22 library. Formatter's zero value and New both show at most ten items.
JSON is a stable consumer format: absent items are null; a known empty list is [].
Add NewWithLimit(limit int) (Formatter, error): limit zero explicitly means no limit;
negative limits are rejected. Keep New's exact func() Formatter type, existing
zero behavior, and JSON tags. Format accepts a borrowed input and must not change
it. Document and test the new configuration and established representations.
