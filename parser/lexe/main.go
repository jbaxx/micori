package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"gg.rocks/lex/capture"
)

type JsonType int

const (
	IllegalJSON JsonType = iota
	JsonArray
	JsonNewline
)

type TokenType int

const (
	IllegalToken TokenType = iota
	SquareBracket
	CurlyBrace
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
	j.content.Write(content)
	return nil
}

func (j *JsonFile) ContentReader() io.Reader {
	return bytes.NewReader(j.content.Bytes())
}

func (j *JsonFile) ContentBytes() []byte {
	return j.content.Bytes()
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func (j *JsonFile) ValidateJSON() error {
	jsontype := j.getJSONType()
	if jsontype == IllegalJSON {
		return fmt.Errorf("Invalid JSON")
	}
	j.filetype = jsontype
	return nil
}

func (j *JsonFile) getJSONType() JsonType {
	t := getFirstToken(j.ContentReader())
	if t == CurlyBrace {
		fmt.Println("Json Newline")
		isValidJSONNewline(j.ContentReader())
		return JsonNewline
	}
	if t == SquareBracket {
		fmt.Println("Json Array")
		isValidJSONArray(j.ContentReader())
		return JsonArray
	}
	return IllegalJSON
}
func isValidJSONNewline(r io.Reader) (bool, error) {

	d := json.NewDecoder(r)

	var m map[string]interface{}

	for {
		var v interface{}
		err := d.Decode(&v)
		if err == io.EOF {
			break
		}

		t, ok := v.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("Invalid JSON Newline")
		}
		m = t
	}
	_ = m
	fmt.Println("Valid Json Newline")
	return true, nil
}

func isValidJSONArray(r io.Reader) (bool, error) {

	d := json.NewDecoder(r)

	var m []interface{}

	for {
		var v interface{}
		err := d.Decode(&v)
		if err == io.EOF {
			break
		}

		t, ok := v.([]interface{})
		if !ok {
			return false, fmt.Errorf("Invalid JSON Array")
		}
		m = t
	}
	_ = m
	fmt.Println("Valid Json Array")
	return true, nil
}

func getFirstToken(r io.Reader) TokenType {
	br := bufio.NewReader(r)
	for {
		ch, _, err := br.ReadRune()
		if err == io.EOF {
			return IllegalToken
		}
		if !isWhitespace(ch) {
			switch ch {
			case '[':
				return SquareBracket
			case '{':
				return CurlyBrace
			default:
				return IllegalToken
			}
		}
	}
}

func (j *JsonFile) Capture() error {
	losbytes, err := capture.CaptureInputFromEditor(j.ContentBytes(), capture.GetPreferredEditorFromEnvironment)
	if err != nil {
		return err
	}
	j.content.Reset()
	j.content.Write(losbytes)
	return nil
}

func (j *JsonFile) String() string {
	return j.content.String()
}

func main() {
	flag.Parse()

	file, err := os.Open(*filename)
	if err != nil {
		log.Fatal(err)
	}

	ff := NewJsonFile()
	ff.ReadContent(file)

	err = ff.ValidateJSON()
	if err != nil {
		log.Fatal(err)
	}

	err = ff.Capture()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ff)

	err = ff.Capture()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ff)

}
