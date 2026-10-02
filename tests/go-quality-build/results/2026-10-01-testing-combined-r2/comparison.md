# Isolation revision 2 integrated outcomes

| Arm | Architecture | Testing | Correctness |
| --- | --- | --- | --- |
| isolation-only | A+ | B | B |
| both (behavior revision 1, isolation revision 2) | A+ | B | A+ |

Both unchanged candidates reconstruct exactly and pass ordinary, race, ten-shuffle and held checks. Both detect the six frozen compiled regressions and an actual automatic-redirect-following regression. Concrete-client default-policy checks now reject an initial 302 without publication, make one request and preserve the borrowed policy; this corrects the first-pass single-arm gap. The isolation-only borrowed-policy removal diagnostic by itself does not follow a redirect and is labelled policy-call observation; a separate forced-following variant fails meaningful assertions. No held probe contributes to candidate-only mutation credit.

The corrected HTTP boundary does not erase new findings. Isolation-only compares arbitrary error interfaces to a context cause and panics for a valid non-comparable custom error; its suite omits that case. The both suite permits exact-200 to all-2xx widening to survive, although current production rejects 206 correctly. Raw independent cards, checks, added probes, candidate-specific mutation patches and exact changed sources are preserved. Independent controller cause checks require no panic, ordered single release and retained independent failure, without prescribing nil versus returned cause. They confirm the same comparison panic in first-pass architecture and revised isolation-only; revised both handles that valid input safely.

These are moderate contained defects/test gaps, not complete lifecycle failure. Behavior owns the exact acceptance-set and supported interface/error counterexamples. Original behavior revision 2 must receive fresh all-case/control/repeat and affected integrated trials; no candidate repair counts as unseen improvement. Isolation revision 2 remains pending separate promotion assessment. First-pass both's A+/A+/A+ result is one bounded observation, not a guarantee for future authors. Earlier individual review launch labels limit grade-comparison blinding; these integrated outcomes use fresh neutral paths/names.

Executed Go1.26.5 darwin/arm64 against Go1.22 module declarations; actual minimum-toolchain, Windows and runtime skill routing remain unverified. Restricted listener failures and approved unchanged-source checks remain raw. Exact inherited model/reasoning metadata are unavailable, limiting reproduction outside this session.
