
# go-media example applications

This folder provides examples of using the `go-media` module:

  * `mediatool` can be used for extracting metadata and artwork in bulk from
    media files.

This package is part of a wider project, `github.com/djthorpe/go-media`.
Please see the [module documentation](https://github.com/djthorpe/go-media/blob/master/README.md)
for more information.

## Building

In order to build any command, following the instructons for installing `ffmpeg` [here](https://github.com/djthorpe/go-media/blob/master/README.md) and then run:
On Macintosh with homebrew, for example:

```bash
[bash] git clone git@github.com:djthorpe/go-media.git
[bash] cd go-media
[bash] make cmd
```

This will place the application in the `build` folder.

## Application: mediatool

```bash
[bash] mediatool -help

Syntax:
   mediatool <flags> command args...

Commands:
  help       
    	Print this help
  version    
    	Print version of command
  metadata   <file>...
    	Print metadata for one or more files
  artwork    <file>...
    	Extract artwork for one or more files
```

