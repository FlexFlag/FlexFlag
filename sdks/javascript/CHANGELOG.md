# Changelog

All notable changes to the FlexFlag JavaScript SDK will be documented in this file.

## [1.0.4] - 2026-01-15

### Fixed
- Fixed React and Vue dependencies to be optional peer dependencies
- Fixed API request format to match backend expectations
- Fixed batch evaluation endpoint and payload structure
- Fixed authentication header handling for API key evaluation

### Added
- Added `evaluateBoolean()` convenience method for boolean flags
- Added `evaluateString()` convenience method for string flags
- Added `evaluateNumber()` convenience method for number flags
- Added `evaluateJSON<T>()` convenience method for JSON flags
- Added separate entry points for React and Vue integrations
- Added comprehensive usage documentation

### Changed
- Moved React integration to `flexflag-client/react` entry point
- Moved Vue integration to `flexflag-client/vue` entry point
- Core SDK now works in backend environments without React/Vue installed
- Improved error handling and fallback mechanisms

### Breaking Changes
- React imports must now use `import from 'flexflag-client/react'` instead of `'flexflag-client'`
- Vue imports must now use `import from 'flexflag-client/vue'` instead of `'flexflag-client'`

## [1.0.3] - 2024-01-14

### Added
- Initial release with React and Vue support
- Core evaluation functionality
- Caching and offline support
- WebSocket and polling connection modes

## [1.0.2] - 2024-01-13

### Added
- Basic SDK functionality
- Flag evaluation with context
- Batch evaluation support
