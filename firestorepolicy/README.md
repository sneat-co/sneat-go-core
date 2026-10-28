# Firestore placement registry

`root-collections.json` is the reviewed placement contract. Placement never
grants caller authorization; database and request policies remain mandatory.

The `extensions` list names registered owners. By default, existing owners have
placement under `ext/{extension}`, `spaces/{space}/ext/{extension}`, and
`users/{user}/ext/{extension}`. The optional `spaceOnlyExtensions` list is a
unique subset of `extensions`: those owners have placement only below
`spaces/{space}/ext/{extension}`. Their global and user-owned paths remain
denied. Sizeus owns `spaces/{space}/ext/sizeus/sizes/{id}` and is space-only.
No root collection or legacy-write permission is added by registration.

This is an additive schema-version-2 codec extension. New readers accept older
artifacts without the optional subset and retain their existing placement.
Older Core codecs reject artifacts containing the unknown field, so consumers
must upgrade Core before accepting the new artifact. The Firebase registry
mirror must use the same bytes and a matching Core version. Registry digests
change when registration changes; consumers with a sealed expected digest
(including the DataTug root monitor) must coordinate that configuration when
they upgrade. Publishing Core does not upgrade or reconfigure deployed writers.
