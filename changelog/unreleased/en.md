# Unreleased

## Added

- Provide component and Core changelog management APIs.

## Changed

- Provide English and Simplified Chinese READMEs with language links, documentation homepages and a shared Views badge; move brand images to `docs/assets`.
- Generate plugin and agent projects without an i18n directory; the example replies with `Hello AlonaBot` directly. Generated Go files and `plugin.yaml` contain no comments. Add translation files when needed.

## Fixed

- Return published Core changelogs in English or Simplified Chinese according to `Accept-Language`, with English fallback. Exclude unreleased changes from the API.
