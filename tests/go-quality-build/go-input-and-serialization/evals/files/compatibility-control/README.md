Update Decode to trim surrounding whitespace from Name after decoding. Keep the API and Go 1.22 minimum; add a focused regression test.

This existing JSON v1 configuration format accepts unknown members, matches field names case-insensitively, and lets the last duplicate scalar member win. Preserve those rules. Encode must keep nil Tags as null and nonnil empty Tags as [], with the tags member always present. No other normalization or validation is requested.
