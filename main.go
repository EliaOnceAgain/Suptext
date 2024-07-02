package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"
    "github.com/eliaonceagain/suptext/src"
)

func main() {
    // Read input file name
    args := os.Args
    if len(args) == 1 {
        log.Fatal("Missing fname input")
    } else if len(args) > 2 {
        log.Fatal(args[0], " supports a single file input")
    }
    fname := args[1]
    log.Printf("Reading SUP file: %s", fname)

    // Open input file
    fin, err := os.Open(fname)
    if err != nil {
        log.Fatal(err)
    }
    defer fin.Close()

    // Create a buffered reader
    r := bufio.NewReader(fin)
    pgs, err := suptext.ReadPGS(r)
    if err != nil {
        log.Fatalf("Failed parsing segment header: %s", err)
    }

    // Open output file SRT
    srt_fname := fmt.Sprintf("%s.srt", strings.TrimSuffix(fname, filepath.Ext(fname)))
    fout, err := os.Create(srt_fname)
    if err != nil {
        log.Fatalf("Failed to create file: %v", err)
    }
    defer fout.Close()

    // Dump SRT
    log.Printf("Writing SRT file: %s", srt_fname)
    pgs.ToSRT(fout)
    log.Println("Success")
}
