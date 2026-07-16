# Changelog

## 0.4.0

- Add slice support — delimited env values populate slices of any supported scalar type (`[]string`, `[]int`, `[]bool`, etc.)
- Add `delim=` tag option to configure the slice element separator (default `,`)
- Validate `choices` per element for slice fields
- Add nested-struct support via the `envPrefix` tag — recurse into `struct`/`*struct` fields with a composed variable-name prefix
- Add the required package card image to the README

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
