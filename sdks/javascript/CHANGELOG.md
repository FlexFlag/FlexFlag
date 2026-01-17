# Changelog

All notable changes to the FlexFlag JavaScript SDK will be documented in this file.

## [1.1.1] - 2026-01-17

### Fixed
- Fixed `evaluateBoolean()` to properly parse string boolean values ("true"/"false")
- Fixed boolean conversion to handle edge cases: numbers (0=false, non-zero=true), null/undefined
- The method now correctly returns `false` for the string "false" instead of `true`

## [1.1.0] - 2026-01-17

### Added
- **NEW**: Full SSE (Server-Sent Events) support with proper API key authentication
- Real-time flag updates via SSE streaming with automatic reconnection
- Support for multiple SSE event types: `connected`, `flag_update`, `ping`
- Proper URL encoding for API keys in query parameters

### Changed
- SSE endpoint now uses `/api/v1/stream` with API key authentication
- Changed default connection mode back to 'streaming' for real-time updates
- Improved SSE error handling and reconnection logic

### Fixed
- Fixed SSE authentication to work with backend API key validation
- Resolved issues with EventSource not supporting custom headers

## [1.0.7] - 2026-01-17

### Changed
- **BREAKING**: Changed default connection mode from 'streaming' to 'polling'
- Polling is now the recommended mode for client SDKs
- SSE streaming mode remains available but is intended for edge server deployments

### Fixed
- Resolved authentication issues by using polling mode which properly supports API key headers
- Improved reliability of flag updates with polling-based change detection

## [1.0.6] - 2026-01-17

### Fixed
- Fixed SSE authentication by passing API key as query parameter instead of header
- EventSource API doesn't support custom headers, so API key is now sent in URL

## [1.0.5] - 2026-01-17

### Changed
- Migrated from WebSocket to Server-Sent Events (SSE) for real-time flag updates
- Updated streaming connection to use EventSource API instead of WebSocket
- Improved connection stability and automatic reconnection handling for SSE

### Fixed
- Better fallback mechanism when SSE is not available
- Improved error handling for streaming connections

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
- WebSocket and polling connection modes (WebSocket later replaced with SSE in v1.0.5)

## [1.0.2] - 2024-01-13

### Added
- Basic SDK functionality
- Flag evaluation with context
- Batch evaluation support
