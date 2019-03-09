dec-decode: an iso.dec decoder written in Go
============================================

This tool implements the NASOS method of decoding .iso.dec files
back into plain .iso files.

History
-------

This tool is also written in Go and born of a frustration that:
1. the NASOS tool is being used to archive ISOs with limited implementations available
2. the original NASOS tool is written as a .NET GUI tool for Windows and is slow on Linux
3. it is hard to find a canonical site for getting the tool
4. It requires downloading unknown binaries 

This is a pure Go implementation that documents how this works and
should work on all platforms supported by Go. This alleviates all the
concerns above.

Installation
------------

You need a working [Go](https://golang.org/) installation (I used Go 1.12 on Ubuntu Linux 18.04)

For Go < 1.11 the you will need to install the required libraries manually:

    go get github.com/jessevdk/go-flags

Build the tool with:

    go install

Usage
-----
    dec-decode [OPTIONS] Files...

    Application Options:
    -s, --suffix=  add a suffix to the output file
    -v, --verbose  show lots more information than is probably necessary

    Help Options:
    -h, --help     Show this help message

    Arguments:
    Files:         list of files to decode

Status
------

System  | Size | Status
--------|------|-------
Wii     | DVD5 | Working
Wii     | DVD9 | Working
GameCube| DISC | Untested

