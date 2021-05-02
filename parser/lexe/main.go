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
	payload  Payload
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
		j.isValidJSONNewline()
		return JsonNewline
	}
	if t == SquareBracket {
		fmt.Println("Json Array")
		j.isValidJSONArray()
		return JsonArray
	}
	return IllegalJSON
}
func (j *JsonFile) isValidJSONNewline() (bool, error) {

	d := json.NewDecoder(j.ContentReader())

	var m map[string]interface{}

	counter := 0
	for {
		counter++
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
	fmt.Println("Counter: ", counter)
	_ = m
	fmt.Println("Valid Json Newline")
	return true, nil
}

func (j *JsonFile) isValidJSONArray() (bool, error) {

	d := json.NewDecoder(j.ContentReader())

	var m []interface{}

	counter := 0
	for {
		counter++
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
	fmt.Println("Counter: ", counter)
	_ = m
	fmt.Println("Valid Json Array")
	return true, nil
}

func (j *JsonFile) ParsePayloads() {

	if j.filetype == JsonArray {
		j.parseJSONArray()
	}
	if j.filetype == JsonNewline {
		j.parseJSONNewline()
	}

}

func (j *JsonFile) parseJSONArray() error {
	d := json.NewDecoder(j.ContentReader())
	for {
		err := d.Decode(&j.payload.payloads)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	j.payload.count = len(j.payload.payloads)
	return nil
}

func (j *JsonFile) parseJSONNewline() error {
	d := json.NewDecoder(j.ContentReader())
	for {
		var v map[string]interface{}
		err := d.Decode(&v)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		j.payload.payloads = append(j.payload.payloads, v)
	}
	j.payload.count = len(j.payload.payloads)
	return nil
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

type Payload struct {
	payloads []map[string]interface{}
	count    int
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

	// err = ff.Capture()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(ff)

	// err = ff.Capture()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(ff)

	ff.ParsePayloads()

}
