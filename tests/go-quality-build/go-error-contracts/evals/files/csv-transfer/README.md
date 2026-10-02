# csvselect
Go 1.22 library. WriteRows writes valid CSV to a borrowed io.Writer; success means
all encoded bytes reached that writer. Underlying writer errors are inspectable.
Add WriteSelected(w, rows, keep) (int,error) using keep func([]string)bool.
It writes only selected rows in input order and returns the selected row count on
success. On failure, the count is the number of selected rows submitted to the CSV
encoder before the observed failure; buffered encoding is not durable per-row
acknowledgment. A nil keep selects everything. Do not close the borrowed writer.
Keep WriteRows behavior and signature. No automatic retries or new dependency.
