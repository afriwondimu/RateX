# RateX - Go Rate Limiter

RateX is a flexible rate limiting package for Go applications with built-in Gin middleware support. It provides multiple algorithms and storage backends for distributed rate limiting.

## Features

- **Multiple Algorithms**: Token Bucket and Sliding Window
- **Distributed**: Redis support for horizontal scaling
- **Flexible Keys**: By IP, API key, user ID, or custom logic
- **Standard Headers**: X-RateLimit headers
- **Production Ready**: Thread-safe, error handling, and testing
- **Easy Integration**: Simple middleware for Gin
- **Configurable**: Per-route or global limits

## Installation

```bash
go get github.com/afriwondimu/RateX