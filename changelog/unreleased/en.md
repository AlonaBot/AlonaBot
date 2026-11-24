# Unreleased

### Fixed

- Once the platform finishes sending or editing a message, `Send` and `Edit` return its receipt and any platform error without waiting for `AfterSend` hooks to finish. Core runs the hooks in the same batch in their established order in the background. See [#328](https://github.com/AlonaBot/AlonaBot/pull/328).
- Time spent in the queue counts against an event's deadline. Events that expire before leaving the queue skip `PreEvent` and handler processing. Other events continue with the time they have left. See [#329](https://github.com/AlonaBot/AlonaBot/pull/329).
- When a command or event handler call fails, Core reports the original error instead of misreporting a missing response as a decoding failure. A valid `STOP` in a failure response still stops later handlers. See [#330](https://github.com/AlonaBot/AlonaBot/pull/330).
- Editing an earlier reply in the same session still keeps its original mentions and prefix. Core no longer accumulates those associations in memory as it sends more messages. See [#331](https://github.com/AlonaBot/AlonaBot/pull/331).
- Plugins and agents that omit `sleep_capable` allow Core to put them to sleep when eligible. An explicit `false` keeps them resident. If an older builder wrote `false` for an omitted value, rebuild the package with the fixed builder to get the default behavior. See [#332](https://github.com/AlonaBot/AlonaBot/pull/332).
- Core and SDK log messages identify the component, action, state change or reason for failure in plain language. JSON log files retain the full structured fields, and older text logs remain available for queries. See [#333](https://github.com/AlonaBot/AlonaBot/pull/333).
