# 💥 Comic Cruncher

[![CircleCI](https://circleci.com/gh/aimeelaplant/comiccruncher.svg?style=shield&circle-token=f3af6bb29cb3d0dbedf644094dc86cb21b2a552f)](https://circleci.com/gh/aimeelaplant/comiccruncher) [![codecov](https://codecov.io/gh/aimeelaplant/comiccruncher/branch/master/graph/badge.svg?token=nPlAJ6Wzct)](https://codecov.io/gh/aimeelaplant/comiccruncher) ![CircleCI](https://img.shields.io/badge/status-WIP-red.svg)

❗️ Currently **WIP**!

This repository contains the backend code responsible for generating all the character and appearance data as well as serving the REST API for Comic Cruncher.

Comic Cruncher started as an idea from my old Emma Frost site that ranks the character's appearances per year (you can check her graph [here](https://emmafrostfiles.com/comics/)). I was curious what the graphs of other characters looked like, so I ran a small Go script on my computer, and then it turned into this huge thing ... so, here's the code for all who are curious.

The purpose of open sourcing this is more for show and tell  💅, and there is a private dependency required to use some of the packages, so don't expect this to work out of the box.

## Getting Started

You can run this locally with or without Docker, but, again, the purpose of open sourcing this is more for show and tell. 💅

You'll need `docker` and `docker-compose` if you wanna make things easy. Otherwise you can run this locally on your computer without Docker (but no documentation is provided for it).

1. Run `GITHUB_ACCESS_TOKEN=X make netrc` to install the private dependency that hosts the library for external issue sources. (The `docker-compose.yml` will mount the file to the `$HOME` directory of the container.)
- _NOTE_: You'll *need* this access token if you plan on touching most of the cerebro package.
2. Run `make docker-up` to create the Docker container.
3. Run `make docker-dep-ensure` to install the dependencies.
4. Run `make docker-migrations` to run the database migrations.
5. If you wanna interact with the packages that use 3rd party sources, then SET YOUR ENVIRONMENT VARIABLES on your local machine (env vars defined in `docker-compose.yml`):
  - `CC_MARVEL_PUBLIC_KEY` and `CC_MARVEL_PRIVATE_KEY`: the access tokens provided by the Marvel.com API.
  - `CC_AWS_ACCESS_KEY_ID`, `CC_AWS_SECRET_ACCESS_KEY`, `CC_AWS_BUCKET`, `CC_CDN_URL`: The AWS access tokens and CDN url for uploading character images to the remote storage facility.
  - `CC_AWS_REGION` and `CC_AWS_SQS_QUEUE`: The AWS SQS queue for the `messaging` package (which I have yet to actually use. Idea was to sync characters across servers). Note that the access tokens from AWS must have permissions for the specified SQS.

## Package Information

### 🎭 Comic

The `comic` package contains the model definitions, repositories, and services for everything that makes up a comic book: publishers, characters, issues, etc - _not_ really a comedian as the emoji suggests!

The persistence layer is backed by a Postgres database, and Redis is used to cache the simple appearances per year for a character.

### 🧠 Cerebro 

`cerebro` is the application that _generates_ all the characters, character issues, character sources, and appearances per year and sends them to our comic persistence! I don't manually create the characters and issues, as that would be tedious work. 

Check the package README for details.

### DC / Marvel

The `dc` and `marvel` packages are just small wrappers for hitting the DC and Marvel APIs and getting a list of available characters. `cerebro` uses this package.

### ✉️ Messaging

Currently a WIP. This uses the AWS SQS and sends a message to a queue when a character should be synced. For now I'm just using a cron job to sync character appearances and will come back to this package later.

### 🔎 Search

The `search` service is used currently for searching characters in our persistence layer. The implementation just uses the trigram similarities for now.

### 🗄 Storage

Stores images into the remote AWS S3 repository.

### 🌐 Web

The `web` package is the REST API with endpoints for characters, their appearances per year, and a character search tool. The Comic Cruncher frontend interacts with this API to serve the frontend.

## Tests

### Running tests

1. Run `make docker-up-test` to create the Docker container for testing.
2. Run `make docker-migrations-test` to create the test database.
3. Run `make docker-test` to run the tests. 

### Writing tests

The test coverage could be improved!

Limit database tests to the repository layers only. Any other tests that use the repositories must be mocked. The package `internal/mocks` contains generated mocks.

You can use `make docker-mockgen` or `make mockgen` (run `mockgen` locally) to regenerate mocks.

## Deployment

Deployment is janky: no Puppet, no Terraform, no Kubernetes -- 10's are for work! Jobs in `.circleci/config.yml` run the tests, build the application binaries, and pop the binaries directly onto the server...

The latest tagged release is production (not master), so a CircleCI build that deploys to production is triggered any time a release occurs.

## Pull Requests

Pull requests are currently closed for non-collaborators. You can read why at `.github/CONTRIBUTING.md`.
