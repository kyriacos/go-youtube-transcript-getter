# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - TBD

### Added

- Library to fetch YouTube video transcripts via Innertube API.
- CLI tool `youtube-transcript-getter` for fetching transcripts from the command line.
- Support for video IDs and various YouTube URL formats.
- Built-in in-memory and file system caching with configurable TTL.
- Pluggable HTTP fetchers (VideoFetch, PlayerFetch, TranscriptFetch) for proxies and custom headers.
- Typed errors: VideoUnavailableError, DisabledError, NotAvailableError, NotAvailableLanguageError, TooManyRequestError, InvalidVideoIdError.
- Configurable language, user agent, and HTTPS behavior.
