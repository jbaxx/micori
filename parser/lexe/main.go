package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gg.rocks/lex/parser"
)

var filename = flag.String("filename", "", "")

func main() {
	flag.Parse()

	file, err := os.Open(*filename)
	if err != nil {
		log.Fatal(err)
	}

	ff := parser.NewJsonFile()
	ff.ReadContent(file)

	err = ff.ValidateJSON()
	if err != nil {
		log.Fatal(err)
	}

	ff.ParsePayloads()

	// err = ff.Capture()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(ff)

	err = ff.CaptureCollection()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ff)

	err = ff.CaptureCollection()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ff)

	// err = ff.Subset(0, 5)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(ff)

	fmt.Println()

	ff.AddVisit("item")
	ff.Visit()
	ff.ResetVisit()

	fmt.Println()

	ff.AddVisit("size")
	ff.AddVisit("h")
	w := func(i interface{}) bool {
		fl := i.(float64)
		fmt.Printf("greater than 10: %v \n", fl > 10)
		return true
	}
	ff.AddFunction(w)
	ff.Visit()
	ff.ResetVisit()

	fmt.Println()

	ff.AddVisit("size")
	ff.AddVisit("uom")
	w = func(i interface{}) bool {
		fmt.Printf("suerte: %v \n", i)
		return true
	}
	ff.AddFunction(w)
	ff.Visit()
	ff.ResetVisit()

	fmt.Println()
	fmt.Println("###")
	fmt.Println()

	ff.AddVisit("size")
	ff.AddVisit("h")
	ff.AddFunction(w)
	ff.Visit()

}
