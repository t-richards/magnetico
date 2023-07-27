# magnetico

*Autonomous (self-hosted) BitTorrent DHT search engine suite.*

This is a hard fork of the original magnetico project, with some opinionated changes.

> [!IMPORTANT]\
> This project is a work in progress.
> It does not intend to be a drop-in replacement for the original magnetico.

## About

**magnetico** is an autonomous (self-hosted) BitTorrent DHT search engine suite that is *designed for end-users*. It crawls the BitTorrent DHT network in the background to discover info hashes and fetch metadata from peers. It also provides a lightweight web interface to search and browse discovered torrents.

This program allows anyone with a decent Internet connection to access the vast amount of torrents waiting to be discovered within the BitTorrent DHT space, *without relying on any central entity*.

**magnetico** liberates BitTorrent from the yoke of centralised trackers & web-sites and makes it
*truly decentralised*. Finally!

## Features

Easy installation & minimal requirements:
 - We provide pre-compiled static binaries for most platforms (currently macOS + Linux).
 - The application runs without root or admin privileges.

## Installation

Download the latest release from the [releases page](https://github.com/t-richards/magnetico/releases).

## Changes from the original project

 - Updated the code for modern Go, making it easier to build and run.
 - Replaced multiple database interfaces with the one true database: SQLite.
 - Removed the JavaScript, because it is ugly and unnecessary.
 - Removed API routes, as they will not be used.
 - Merged the crawler and web server into a single binary.
 - Many bugfixes and code quality-of-life improvements.

## Roadmap

 - [x] Fix database code
   - [x] Instant corruption when applying schema
   - [x] Analyze and flatten migrations
   - [ ] Maybe replace hand-rolled queries with an ORM?
 - [x] Replace client-side handlebars templates with go templates
 - [x] Add robots noindex/nofollow headers
 - [ ] Fix pagination
 - [ ] Pretty up the files list
 - [ ] Fix crawler code
   - [ ] `Could NOT write an UDP packet! invalid argument`
 - [ ] Automate docker image building
 - [ ] Windows support? (PRs welcome)

## Why?

BitTorrent, being a distributed P2P file sharing protocol, has long suffered because of the
centralised entities that people depended on for searching torrents (websites) and for discovering
other peers (trackers). Introduction of DHT (distributed hash table) eliminated the need for
trackers, allowing peers to discover each other through other peers and to fetch metadata from the
leechers & seeders in the network. **magnetico** is the finishing move that allows users to search
for torrents in the network, hence removing the need for centralised torrent websites.

## License

magnetico is licensed under the [GNU AGPLv3](./COPYING).
