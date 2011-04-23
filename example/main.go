package main

import "flag"
import "gob"
import "os"

var indexPath *string = flag.String("index", "", "path to the index file")
var createIndex *bool = flag.Bool("create", false, "create a new index")

func loadIndex() *Index {
	var index *Index = nil

	if *createIndex {
		index = NewIndex()
	} else {
		indexReader, err := os.Open(*indexPath, 0, 0)

		if err != nil{
			panic(err)
		}

		index = new(Index)
		err = gob.NewDecoder(indexReader).Decode(&index)

		if err != nil{
			panic(err)
		}
	}

	return index
}

func main() {
	flag.Parse()
	index := loadIndex()

	index.Lookup("foo")
}
