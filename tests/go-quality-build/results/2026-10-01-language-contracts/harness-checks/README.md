# Private-check helper regression validation

A controller-created copy of the passing calculation output was deliberately
changed to `return items/size+1`. The helper ran real uncached race-enabled author
checks, vet, and private probes. Both test commands detected the regression;
`check(...)` returned False. The driver now exits nonzero for a false result.
This validates failure signaling; it is not an agent outcome or candidate uplift.
The raw result is preserved in verification-negative.json. No trial output or
canonical fixture was mutated. Go 1.26.5, GOWORK=off and writable private cache.
