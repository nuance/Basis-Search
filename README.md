Basis Search
============

A bunch of search data structures implemented in go. This isn't complete - it's mostly just posting lists + posting list query code, at the moment, but I hope to turn this into a library of building blocks, like Lucene, but in a sane language and with a much tighter focus.

Design Notes
------------

Search services need a few basic things.

 * Querying - Queries should gather posting lists from the index, drive the intersect / merge cycle, fetch the scored documents, and handle gathering final results. This portion is considerably less flexible than Lucene, since it doesn't need to be (I'll assume you'll be building your own service, and thus can build the complexity where you need it).

 * Matching - Fast intersections / merges. For text this is a trie or hash table pointing to posting lists. Discrete-valued attributes should probably be B+ trees. Geo data should probably be some kind of tree (r-tree or quad tree).

 * Scoring - Query needs to be able to fetch a score doc, from the index, stored as gobs. These should implement a score method (they could be just comparable, like Lucene does it, but it's easier to think about each doc as having a floating point score).

 * Indexing - Load an existing index and expand it, then serialize it back to disk. Indexing here is assumed to happen separately from (most) of your searching. Maybe you want to perform realtime updates of all your nodes, but you probably don't want to flush them to disk. Index updates should be atomic, but not transactional (you'll pay way too much for this).

There are things that they don't need.

 * Language Analysis - Lucene does this. Well. But it's still a mistake. If you need it, build an analysis service on lucene, or use NLTK w/ python. If a good library is written for go, then use it, but it's definitely out of the scope of this project (and extremely useful as a standalone tool).

Code Layout
-----------

    src/ - the main source for the library
      util/ - Utility functions (variable-sized ints, buffer pool etc.)
	  match/ - Structures that store matches (posting lists, bitsets) and algorithms for merging / intersecting them 
      index/ - Structures for looking up matches (trie, quad tree, b+ tree)

    example/ - an example of a simple text + attribute search server
