
# go-media chromaprint bindings

This package provides bindings for [chromaprint](https://acoustid.org/chromaprint) audio fingerprinting.

This package is part of a wider project, `github.com/djthorpe/go-media`.
Please see the [module documentation](https://github.com/djthorpe/go-media/blob/master/README.md)
for more information.

## Building

In order to use this package, you will need to install the chromaprint libraries. 
On Darwin (Mac) with Homebrew installed:

```bash
[zsh] brew install chromaprint
[zsh] go get git@github.com:djthorpe/go-media.git
[zsh] cd go-media/sys/chromaprint
```

For Linux Debian,

```bash
[bash] sudo apt install libchromaprint-dev
[zsh] go get git@github.com:djthorpe/go-media.git
[zsh] cd go-media/sys/chromaprint
```

For more information:

  * API Documentation Sources: https://github.com/acoustid/chromaprint
  * Web Service: https://acoustid.org/webservice

This package provides low-level library bindings. There is also a 
[client library for the web service](https://github.com/djthorpe/go-media/tree/master/pkg/chromaprint).

## Usage

TODO


