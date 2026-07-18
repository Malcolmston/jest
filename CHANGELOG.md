# Changelog

All notable changes to this project are documented in this file. The format is
based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-07-18

Another step toward Jest parity. Still pure Go standard library, no third-party
dependencies.

### Added

- **Truthiness matchers**: `Matcher.ToBeTruthy` and `Matcher.ToBeFalsy` using
  JavaScript-style truthiness (nil, `false`, numeric zero, `""` and NaN are
  falsy), plus `Matcher.ToBeNull` as the Jest spelling of a nil check.
- **Collection matchers**: `Matcher.ToContainEqual` (asymmetric-aware deep
  element/value membership), `Matcher.ToHaveLength` (Jest spelling of
  `ToHaveLen`) and `Matcher.ToBeOneOf`.
- **Return-value matchers**: `Matcher.ToHaveReturnedTimes`,
  `Matcher.ToHaveLastReturnedWith` and `Matcher.ToHaveNthReturnedWith`.
- **Jest alias spellings**: `Matcher.ToBeCalled`, `ToBeCalledTimes`,
  `ToBeCalledWith`, `LastCalledWith`, `NthCalledWith`, `ToReturn` and
  `ToReturnWith`.
- **Inline & thrown-error snapshots**: `Matcher.ToMatchInlineSnapshot` compares
  against an inline serialized literal, and `Matcher.ToThrowMatchingSnapshot`
  snapshots a captured panic message (mirroring `toThrowErrorMatchingSnapshot`).
- **Asymmetric matchers**: `CloseTo` (mirrors `expect.closeTo`) and the negated
  forms `NotArrayContaining`, `NotObjectContaining`, `NotStringContaining` and
  `NotStringMatching` (mirroring `expect.not.*`).
- **Mock lifecycle**: `Mock.MockClear`, `Mock.MockReset`, `Mock.MockName`, and
  the global `ClearAllMocks` / `ResetAllMocks` backed by a registry that every
  `NewMock` (and therefore every `Fn*`, `Spy*` and `SpyOn`) joins.
- **Fake-timer controls**: `Clock.AdvanceTimersToNextTimer`,
  `Clock.ClearAllTimers`, `Clock.GetTimerCount` and `Clock.SetSystemTime`.

## [0.2.0] - 2026-07-17

A large step toward Jest parity. Still pure Go standard library, no third-party
dependencies.

### Added

- **Snapshot testing**: `Matcher.ToMatchSnapshot(name...)` with an on-disk store
  under `__snapshots__/`, a stable deterministic serializer, and an update mode
  via the `JEST_UPDATE_SNAPSHOTS` environment variable or `SetUpdateSnapshots`.
- **Asymmetric matchers** usable inside `ToEqual`, `ToMatchObject`,
  `ToHaveProperty` and the call-argument matchers: `Any(type)`, `Anything()`,
  `StringContaining`, `StringMatching`, `ArrayContaining`, `ObjectContaining`.
- **New value matchers**: `ToMatchObject`, `ToStrictEqual`, `ToHaveProperty`
  (dotted / indexed paths), `ToBeInstanceOf`, `ToBeDefined`, `ToBeUndefined`,
  `ToBeNaN`.
- **Mock-oriented matchers**: `ToHaveBeenCalled`, `ToHaveBeenCalledTimes`,
  `ToHaveBeenCalledWith`, `ToHaveBeenNthCalledWith`, `ToHaveBeenLastCalledWith`,
  `ToHaveReturned`, `ToHaveReturnedWith`.
- **Fake timers**: `Clock` with `SetTimeout`, `SetInterval`, `ClearTimer`,
  `After`, `AdvanceTimersByTime`, `RunAllTimers`, `RunOnlyPendingTimers`, `Now`.
- **Mock enrichment**: `MockImplementation`, `MockImplementationOnce`,
  `MockReturnValueOnce`, `MockResolvedValue`, `MockRejectedValue`, `Results`,
  and per-call panic tracking (`Call.Panicked`).
- **Spying**: `SpyOn` replaces a function variable or struct field in place
  (reflection-based, records calls, delegates to the original), plus
  `Spy.Restore` and `RestoreAllMocks`.
- **Suite lifecycle**: `BeforeAll` / `AfterAll` in addition to the existing
  `BeforeEach` / `AfterEach`.
- **Parameterized tests**: `Each` and `DescribeEach`; plus `ItSkip`, `ItOnly`
  and `ItTodo`.
- **Custom matchers**: `Extend` registers named matchers invoked via `Matcher.To`.
- **Assertion counting**: `Assertions(t, n)` and `HasAssertions(t)`.

### Changed

- `Matcher.ToEqual` now performs asymmetric-aware deep equality; behavior is
  unchanged for values without asymmetric matchers.

## [0.1.0]

- Initial release: fluent expectations, mocks/spies, and `Describe`/`It`
  organization over Go's `testing` package.
