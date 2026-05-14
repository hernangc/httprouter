# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- **ApplicationGroup**: Tree-like structure of hierarchical route groups

### Changed
- Many http.HandlerFunc uses changed for http.Handler 


## [0.2.2] - 2025-06-18
### Changed
- Updated module name

## [0.2.1] - 2025-06-17
### Fixed
- Fixed module name

## [0.2.0] - 2025-06-17
### Added
- Support for global Middleware.
- Added unit and bench tests.

## [0.1.0] - 2025-06-15

### Added
- Initial release of the HttpRouter package
- HTTP method-based routing with support for GET, POST, PUT, PATCH, and DELETE methods
- Middleware support with composition capabilities
- Automatic 405 Method Not Allowed responses with proper Allow headers
- Integration with standard Go `http.ServeMux`
- **WebApplication Interface**:
  - Core interface defining HTTP method routing functions
  - Support for middleware chains on any route
  - Application struct implementing the WebApplication interface 
  - Middleware composition functionality 
  - Method-specific handler registration
