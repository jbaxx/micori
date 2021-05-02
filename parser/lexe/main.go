package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type JsonType int

const (
	IllegalJSON JsonType = iota
	JsonArray
	JsonNewline
)

var filename = flag.String("filename", "", "")

type JsonFile struct {
	content  *bytes.Buffer
	filetype JsonType
}

func NewJsonFile() *JsonFile {
	b := new(bytes.Buffer)
	return &JsonFile{
		content: b,
	}
}

func (j *JsonFile) ReadContent(r io.Reader) error {
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	j.content.Read(content)
	return nil
}

func (j *JsonFile) ContentReader() io.Reader {
	return bytes.NewReader(j.content.Bytes())
}

func (j *JsonFile) ContentBytes() []byte {
	return j.content.Bytes()
}

func main() {
	flag.Parse()

	file, err := os.Open(*filename)
	if err != nil {
		log.Fatal(err)
	}

	ff := NewJsonFile()
	ff.ReadContent(file)

	fmt.Println(ff.content)

	fmt.Println(ff.ContentBytes())
}
