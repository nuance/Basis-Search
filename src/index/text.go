package index

import "os"
import posting "gq/posting"

type TextIndex struct {
	Docs map[string] *posting.PostingList
	Deleted map[int] bool

	Path string
	DocsBacking *os.File
	DeletedBacking *os.File
}

type TextResult struct {
	Id int
	Score float64
}

func (idx *TextIndex) Query(tokens []string) <-chan TextResult {
	results := make(chan TextResult)

	go func (out chan<- TextResult) {
		defer close(out)

		pls := make([]*posting.PostingList, len(tokens))
		dfs := make([]float64, len(tokens))

		for _, word := range tokens {
			if idx.Docs[word] == nil {
				return
			}

			pls = append(pls, idx.Docs[word])
			dfs = append(dfs, float64(idx.Docs[word].Stats().DocCount))
		}

		for docs := range posting.Intersection(pls) {
			tfidf := 0.0

			for idx, tf := range docs.Payloads {
				tfidf += float64(tf) / float64(dfs[idx]) 
			}

			out <- TextResult{docs.Doc, tfidf}
		}
	} (results)

	return results
}
