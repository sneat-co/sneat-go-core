# Firestore placement registry

`root-collections.json` is the reviewed placement contract. Placement never
grants caller authorization; database and request policies remain mandatory.

The `extensions` list names registered owners. By default, existing owners have
placement under `ext/{extension}`, `spaces/{space}/ext/{extension}`, and
`users/{user}/ext/{extension}`. The optional `spaceOnlyExtensions` list is a
unique subset of `extensions`: those owners have placement only below
`spaces/{space}/ext/{extension}`. Their global and user-owned paths remain
denied. Sizeus owns `spaces/{space}/ext/sizeus/sizes/{id}` and is space-only.
Emailing (the sneat-core-modules queued-email pipeline) owns the server-only
records at `ext/emailing/emails/{id}`; the former root `EmailAddress` stays
unapproved.
AskStage is registered with default placement, not space-only: besides its
Space records under `spaces/{space}/ext/askstage`, it owns two global paths
outside any Space, the public show note `ext/askstage/showStamps/{stampID}` and
the short-address claims `ext/askstage/slugNamespaces/show/claims/{code}`.
Videostreaming is registered with default placement, not space-only: it owns
the global channel records at
`ext/videostreaming/platforms/{platform}/channels/{channelID}` with
`destinations` (encrypted stream keys), `backgrounds` and `streamEvents` below.
Channels are global by design (founder ruling, 2026-11-19: "not at space");
the registry has no global-only mode, so its Space and user placement is
allowed but unused. Its records are server-only in the Firebase rules.
No root collection or legacy-write permission is added by registration.

This is an additive schema-version-2 codec extension. New readers accept older
artifacts without the optional subset and retain their existing placement.
Older Core codecs reject artifacts containing the unknown field, so consumers
must upgrade Core before accepting the new artifact. The Firebase registry
mirror must use the same bytes and a matching Core version. Registry digests
change when registration changes; consumers with a sealed expected digest
(including the DataTug root monitor) must coordinate that configuration when
they upgrade. Publishing Core does not upgrade or reconfigure deployed writers.

## Retired observation names

The post-cleanup registry retains six operational observation names. Retired
names are no longer `Legacy`: if recreated, `CompareRoots` reports them as
`Unexpected`, which the root monitor evaluates. This does not grant writes to
the retained names; `legacyWritableRoots` remains empty. Activate this cohort
only after separately approved cleanup and preservation verification. Historical
count snapshots remain immutable evidence; registry retirement does not delete
data or establish writer absence. Consumers must pair the new registry digest
with their selected Core artifact and reviewed monitor configuration.
