package parser

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"

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

type JsonFile struct {
	content    *bytes.Buffer
	filetype   JsonType
	Collection Collection
	visitor    *Visitor
}

type Collection struct {
	payloads    []map[string]interface{}
	payloadsmap map[int]map[string]interface{}
	yielder     map[int]bool
	count       int
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
	j.content.Reset()
	j.content.Write(content)
	return nil
}

func (j *JsonFile) ContentReader() io.Reader {
	return bytes.NewReader(j.content.Bytes())
}

func (j *JsonFile) ContentBytes() []byte {
	return j.content.Bytes()
}

func (j *JsonFile) CollectionBytes() []byte {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	// enc.SetIndent("", "    ")
	for i := 0; i < len(j.Collection.yielder); i++ {
		v, ok := j.Collection.yielder[i]
		if ok && v { // if exists and yields true
			p, _ := j.Collection.payloadsmap[i]
			if err := enc.Encode(p); err != nil {
				log.Println(err)
			}
		}
	}
	return buf.Bytes()
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
		err := d.Decode(&j.Collection.payloads)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	j.Collection.payloadsmap = make(map[int]map[string]interface{})
	j.Collection.yielder = make(map[int]bool)
	for index, payload := range j.Collection.payloads {
		j.Collection.payloadsmap[index] = payload
		j.Collection.yielder[index] = true
	}
	j.Collection.count = len(j.Collection.payloads)
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
		j.Collection.payloads = append(j.Collection.payloads, v)
	}
	j.Collection.payloadsmap = make(map[int]map[string]interface{})
	j.Collection.yielder = make(map[int]bool)
	for index, payload := range j.Collection.payloads {
		j.Collection.payloadsmap[index] = payload
		j.Collection.yielder[index] = true
	}
	j.Collection.count = len(j.Collection.payloads)
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

func (j *JsonFile) CaptureCollection() error {
	losbytes, err := capture.CaptureInputFromEditor(j.CollectionBytes(), capture.GetPreferredEditorFromEnvironment)
	if err != nil {
		return err
	}
	j.ReadContent(bytes.NewReader(losbytes))
	// j.content.Reset()
	// j.content.Write(losbytes)
	return nil
}

func (j *JsonFile) Subset(start, end int) error {
	payloads := j.Collection.payloads
	if end > len(payloads) {
		return fmt.Errorf("Out of index")
	}
	p := payloads[start:end]
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	losbytes, err := capture.CaptureInputFromEditor(b, capture.GetPreferredEditorFromEnvironment)
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

type Visitor struct {
	key  string
	eval func(interface{}) bool
	next *Visitor
}

func NewVisitor() *Visitor {
	return &Visitor{}
}

func (v *Visitor) addVisit(key string) {
	if v.key == "" {
		v.key = key
		return
	}
	if v.next == nil {
		v.next = &Visitor{
			key: key,
		}
		return
	}
	v.addVisit(key)
}

func (j *JsonFile) ResetVisit() {
	j.visitor = nil
}

func (j *JsonFile) AddVisit(key string) {
	if j.visitor == nil {
		j.visitor = NewVisitor()
	}
	j.visitor.addVisit(key)
}

func (v *Visitor) addFunc(f func(interface{}) bool) {
	fmt.Println("adding")
	if v.next == nil {
		v.eval = f
		return
	}
	v.next.addFunc(f)
}

func (j *JsonFile) AddFunction(f func(interface{}) bool) {
	if j.visitor == nil {
		j.visitor = NewVisitor()
	}
	j.visitor.addFunc(f)
}

func (m *Collection) Visit(v *Visitor) {
	for _, payload := range m.payloads {
		value, ok := payload[v.key]
		fmt.Printf("%s", v.key)
		// target in first level
		if ok && v.next == nil {
			fmt.Printf(": %v\n", value)
			if v.eval != nil {
				v.eval(value)
			}
		}
		// still levels to visit
		if ok && v.next != nil {
			valueAsMap, _ := value.(map[string]interface{})
			val, err := keepVisiting(valueAsMap, v.next)
			if err != nil {
				fmt.Println(err)
			}
			// do something with val returned
			_ = val
			// fmt.Printf(": %v\n", val)
		}
		if !ok {
			fmt.Printf("key not found: %s", v.key)
		}
	}
}

func (j *JsonFile) Visit() {
	j.Collection.Visit(j.visitor)
}

func keepVisiting(m map[string]interface{}, v *Visitor) (interface{}, error) {
	fmt.Printf(".%s", v.key)
	if v.next == nil {
		val, ok := m[v.key]
		if !ok {
			return nil, fmt.Errorf("key not found: %s", v.key)
		}
		fmt.Printf(": %v\n", val)
		if v.eval != nil {
			v.eval(val)
		}
		return val, nil
	}

	val, ok := m[v.key]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", v.key)
	}

	nm, ok := val.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("can't proceed with key %s", v.key)
	}

	return keepVisiting(nm, v.next)
}
