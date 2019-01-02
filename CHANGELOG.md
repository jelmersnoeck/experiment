# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/). Changelog

## Unreleased

## v2.1.0 - 2019-01-02

### Added

- `LogPublisher`, a Publisher that logs out the Observation values.

## v2 - 2017-11-15

### Changes
- The experiment engine has been completely rewritten. It's a lot simpler in usage and architecture.

## v1.1.0 - 2016-08-22

### Removed
- Removed testify dependency
- Removed internal interface dependencies, rely on structs from now on
- Removed `x/net/context` dependency

### Changes

- Copy in context interface for backwards compatibility.

## v1.0.0 - 2016-08-08

### Added
- First official release.
