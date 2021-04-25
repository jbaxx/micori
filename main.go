package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

type Schema struct {
	BsonType   string             `json:"bsonType"`
	Title      string             `json:"title"`
	Required   []string           `json:"required"`
	Properties map[string]*Schema `json:"properties"`
}

func NewSchema() *Schema {
	var s Schema
	return &s
}

func (s *Schema) LoadSchema(r io.Reader) {

	stage := "Schema load"
	log.WithFields(log.Fields{"stage": stage}).Info("Loading schema...")

	schema, err := ioutil.ReadAll(r)
	if err != nil {
		log.WithFields(log.Fields{"stage": "Loader"}).Fatal(err)
	}

	err = json.Unmarshal(schema, &s)
	if err != nil {
		log.WithFields(log.Fields{"stage": stage}).Fatal(err)
	}

	log.WithFields(log.Fields{"stage": stage}).Info("Validating schema...")
	s.validateSchema()
	log.WithFields(log.Fields{"stage": stage}).Info("Schema validated.")
}

func (s *Schema) validateSchema() {

	for _, r := range s.Required {
		log.WithFields(log.Fields{"object": s.Title, "field": r}).Debug("Validating required field")
		_, ok := s.Properties[r]
		if !ok {
			log.WithFields(log.Fields{"object": s.Title, "field": r}).Warn("Missing required field")
		}
	}

	if len(s.Properties) != 0 {
		for _, v := range s.Properties {
			v.validateSchema()
		}
	}

}

func init() {
	// set log formatter
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{})

	// set log output
	log.SetOutput(os.Stdout)

	// set log level
	// log.SetLevel(log.DebugLevel)
	log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.WarnLevel)
}

func main() {
	f, err := os.Open("./schemas/school_format.json")
	if err != nil {
		log.Fatal(err)
	}

	s := NewSchema()
	s.LoadSchema(f)

}
