# RateX

A rate limiter for Go apps with Gin support.

## What's inside

- Token bucket algorithm
- Redis support (if you need multiple servers)
- Rate limit by IP, API key, or whatever you want
- Standard rate limit headers

## Install

```bash
go get github.com/afriwondimu/RateX
```
Install
```bash
go get github.com/gin-gonic/gin
go get github.com/redis/go-redis/v9
go get golang.org/x/time/rate
```
## Test
```bash
make run
```
Try 5 request within 10s
```
http://localhost:8080/api/ping
```
