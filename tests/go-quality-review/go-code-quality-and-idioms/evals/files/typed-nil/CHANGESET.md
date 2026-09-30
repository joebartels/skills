# Changeset

ValidateLimit now returns a shared *LimitError variable at the end instead of returning a literal nil on its success path. Callers reject non-nil errors. Valid limits are 1 through 100; this is a local configuration validator, not the entire application.
