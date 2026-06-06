# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Silverfish-Backend is a Go HTTP API that crawls Chinese novel and comic sites, stores results in MongoDB, and serves them through an authenticated REST API. It uses `gorilla/mux` for routing, `goquery` for HTML scraping, and `go-rod` (headless Chromium) for sites that require JS execution.

## Build, run, develop

```bash
# Bring up MongoDB + mongo-express for local dev (mongo on 127.0.0.1:17031)
docker-compose -f docker-compose/dev.yml up -d

# Build / run the backend (requires Go 1.25)
go build -o silverfish-backend
./silverfish-backend                  # reads env from process; or set CONFIG_PATH=./.env

go run .                              # run directly without building

# Production stack (builds the Dockerfile, mounts ../prod.config.json)
docker-compose -f docker-compose/prod.yml up -d --build
```

No test suite exists in the repo (no `*_test.go` files). There is no lint config beyond `go vet` / `gofmt`.

`RECAPTCHA_KEY` is **required** — startup calls `logrus.Fatal` if it is empty. See `.env.example` and `config.go` for the full env surface (`DEBUG`, `PORT`, `HASH_SALT`, `ALLOW_ORIGINS` as a JSON array, `CRAWL_DURATION`, `SSL`/`SSL_PEM`/`SSL_KEY`, `DB_HOST`). When `CONFIG_PATH` is set, `godotenv` loads that file at boot.

On first boot, if the `user` collection is empty, an `admin` account is auto-created with a random 12-char password printed to the log — grab it from stdout.

## Architecture

The code is organized in three layers, plus a router layer on top:

```
main.go                                  wiring + MongoDB session + CORS + HTTP listen
router/                                  HTTP layer (gorilla/mux "blueprints")
  router.go                              composes Auth/Admin/User/API blueprints
  api/api.go + api/v1/{novel,comic}.go   versioned public API under /api/v1
silverfish/                              domain layer
  silverfish.go                          builds Auth/Admin/User/Novel/Comic and wires fetchers
  {auth,admin,user,novel,comic}.go       per-domain services
  entity/                                MongoDB-backed models + MongoInf wrapper
  interface/fetcher.go                   INovelFetcher / IComicFetcher contracts
  usecase/fetcher_*.go                   one file per source site, embed fetcher_base
```

### Request flow

`main.go` opens a single `mgo.Session`, gets three collections (`user`, `novel`, `comic`), wraps each in an `entity.MongoInf`, and hands them to `silverfish.New`. `silverfish.New` constructs the per-domain services AND the maps of site-specific fetchers (`novelFetchers` / `comicFetchers`, keyed by hostname). Those services are passed to `router.NewRouter`, which assembles blueprints — each blueprint owns a URL prefix and registers handlers on a `mux.Subrouter`. Adding/removing a source site = editing the maps in `silverfish/silverfish.go` and adding a `fetcher_<site>.go` in `silverfish/usecase/`.

### Fetcher pattern

Every source site implements `INovelFetcher` or `IComicFetcher` (`silverfish/interface/fetcher.go`). Implementations embed `usecase.Fetcher` (`fetcher_base.go`) for shared helpers:

- `Match(url)` — host matching used to pick a fetcher for an incoming URL.
- `FetchDoc` / `FetchDocWithEncoding` — plain HTTP+goquery; the encoding variant handles GBK/GB18030/Big5 sites.
- `GenerateRodBrowser` — headless Chromium for JS-heavy sites. Hard-coded path `/usr/bin/chromium` — the Dockerfile installs the `chromium` Alpine package to satisfy this. Local dev outside Docker requires Chromium at that path or the binary path will need adjusting.
- `GenerateID` — md5 → base62, first 7 chars; used as the stable per-novel/comic ID.

`Novel.GetNovelByID` / `Comic.GetComicByID` check `LastCrawlTime` against `CrawlDuration` (minutes) and re-crawl through the matching fetcher when stale, then persist via `MongoInf.Update`.

### Auth

`silverfish/auth.go` keeps sessions in an in-memory `map[string]*Session` (not in MongoDB), so **restarting the process logs everyone out**. `ExpireLoop` sweeps expired sessions on each insert. Passwords are SHA-512 with `HASH_SALT`; session tokens are SHA-512 of `account + time + sessionSalt` (sessionSalt is hard-coded to `"SILVERFISH"`). Login endpoints verify Google reCAPTCHA v2 via `Router.VerifyRecaptcha`.

### CORS

`AllowOrigin` is parsed from `ALLOW_ORIGINS` as a **JSON array string** (e.g. `'["https://silverfish.cc"]'`), not a comma-separated list. Allowed methods are fixed to `GET/POST/DELETE/OPTIONS`; only the `Authorization` header is allowed.
