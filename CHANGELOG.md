# Changelog

## 0.3.3

- Standardize README to 3-badge format with emoji Support section
- Update CI checkout action to v5 for Node.js 24 compatibility
- Add GitHub issue templates, dependabot config, and PR template

## 0.3.2

- Consolidate README badges onto single line

## 0.3.1

- Add badges and Development section to README

## 0.3.0

- Fix integer overflow for sized int/uint types (`int8`, `int16`, `uint8`, etc.) by parsing with correct bit size
- Add `encoding.TextUnmarshaler` support for custom types
- Improve error messages to include actual type name

## 0.2.0

- Fix `choices` tag to trim whitespace from individual values
- Validate default values against `choices` constraint
- Add comprehensive test suite

## 0.1.0

- Initial release
