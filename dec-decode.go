package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	Verbose    bool `short:"v" long:"verbose" description:"show lots more information than is probably necessary"`
	Positional struct {
		Files []string `description:"list of files to decode" required:"true"`
	} `positional-args:"true" required:"true"`
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	for _, filePath := range opts.Positional.Files {
		fin, err := os.Open(filePath)
		errorExit(err)
		defer fin.Close()

		fileName := path.Base(filePath)
		var outPath string
		if path.Ext(fileName) == ".dec" {
			outPath = path.Join(".", strings.TrimRight(fileName, ".dec"))
		} else {
			outPath = path.Join(".", fileName+".iso")
		}
		//outPath += ".dmp"

		signature := readSignature(fin)
		fin.Seek(0, io.SeekStart)
		switch signature {
		case "GCML":
		case "GCMM":
			decodeGameCube(fin, outPath)
		case "WII5":
			decodeWii(fin, outPath, 0x1182800) //18360320
		case "WII9":
			decodeWii(fin, outPath, 0x1FB5000) //33247232
		default:
			fmt.Printf("Unknown filetype: %s when checking file %s\n", signature, fileName)
		}
	}
}

func readSignature(r io.Reader) string {
	buffer := make([]byte, 4)
	_, err := io.ReadFull(r, buffer)
	errorExit(err)
	return string(buffer)
}
