# Changelog

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
