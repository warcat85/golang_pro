package hw03frequencyanalysis

import (
	"container/heap"
	"regexp"
	"sort"
	"strings"
)

const (
	ALL = -1
)

var (
	/*
	   Regex to find delimiters
	   Can be:
	   - any number of spaces (possibly having a punctuation character before or after),
	   followed by a punctuation character that follows any number of spaces any number of times
	   (to eliminate single delimiters)
	   - a string starting with punctuation characters
	   - a string ending with punctuation characters

	   Currently this regex is unused, because without lookahead and lookbehind it is impossible
	   to detect a difference if punctuation characters starts straight after the word or the space -
	   therefore it is impossible to detect multiple punctuation characters as a word
	*/
	DELIMITERS = regexp.MustCompile(
		`(\p{P}*(\s+(\p{P}\s+)*|$)|^)\p{P}*`)
	/*
	   Regex to find words
	   Can be:
	   - anything starting by not a punctuation character and not a space, which may follow
	   by any number of not spaces and end by not a punctuation character and not a space
	   - more than one punctuation character
	*/
	DELIMITED = regexp.MustCompile(
		`[^\p{P}\s](?:[^\s]*[^\p{P}\s])*|[\p{P}]{2,}`)

	/*
	   Regex to find word within the string split by words
	   Can be:
	   - anything starting by not a punctuation character, which may follow
	   by any number of characters and end by not a punctuation character
	   - string containing only punctuation characters
	*/
	WORD = regexp.MustCompile(`[^\p{P}](?:.*[^\p{P}])?|^\p{P}{2,}$`)
)

/*
Simple solution:
Reads data, counts occurrences, sorts, gets the first 10 elements.
*/
func Top10Simple(text string) []string {
	wordCount := readCountFields(text)
	return top10AlgoSimple(wordCount)
}

func Top10SimpleRegex(text string) []string {
	wordCount := readCountRegex(text)
	return top10AlgoSimple(wordCount)
}

func Top10SimpleFieldsRegex(text string) []string {
	wordCount := readCountFieldsRegex(text)
	return top10AlgoSimple(wordCount)
}

type Word struct {
	word  string
	count uint
	index int
}

/*
Solution that uses array as the heap for top occurrences

Puts data on the heap as we read it, moves word up if the count increases.
*/
func Top10ArrayHeap(text string) []string {
	wordCount := make(map[string]*Word)
	words := make([]*Word, 0, 10)
	for _, word := range strings.Fields(text) {
		cur, ok := wordCount[word]
		if !ok {
			cur = &Word{word, 0, -1}
			wordCount[word] = cur
		}
		cur.count++
		words = add(words, cur)
	}
	return createResult(words)
}

func Top10ArrayHeapRegex(text string) []string {
	wordCount := make(map[string]*Word)
	words := make([]*Word, 0, 10)
	for _, match := range DELIMITED.FindAllString(text, ALL) {
		word := strings.ToLower(match)
		cur, ok := wordCount[word]
		if !ok {
			cur = &Word{word, 0, -1}
			wordCount[word] = cur
		}
		cur.count++
		words = add(words, cur)
	}
	return createResult(words)
}

func Top10ArrayHeapFieldsRegex(text string) []string {
	wordCount := make(map[string]*Word)
	words := make([]*Word, 0, 10)
	for _, src := range strings.Fields(text) {
		if word, isWord := transformWord(src); isWord {
			cur, ok := wordCount[word]
			if !ok {
				cur = &Word{word, 1, -1}
				wordCount[word] = cur
			} else {
				cur.count++
			}
			words = add(words, cur)
		}
	}
	return createResult(words)
}

/*
Solution that uses array as the heap for top occurrences

First reads all the data and counts, then puts into the heap and counts results.
*/
func Top10PostArrayHeap(text string) []string {
	wordCount := readWordsFields(text)
	return top10AlgoArrayHeap(wordCount)
}

func Top10PostArrayHeapRegex(text string) []string {
	wordCount := readWordsRegex(text)
	return top10AlgoArrayHeap(wordCount)
}

func Top10PostArrayHeapFieldsRegex(text string) []string {
	wordCount := readWordsFieldsRegex(text)
	return top10AlgoArrayHeap(wordCount)
}

/*
Solution that uses MinHeap for top occurrences

First reads all the data and counts, then puts into the heap and counts results.
*/
func Top10PostMinHeap(text string) []string {
	wordCount := readCountFields(text)
	return top10AlgoMinHeap(wordCount)
}

func Top10PostMinHeapRegex(text string) []string {
	wordCount := readCountRegex(text)
	return top10AlgoMinHeap(wordCount)
}

func Top10PostMinHeapFieldsRegex(text string) []string {
	wordCount := readCountFieldsRegex(text)
	return top10AlgoMinHeap(wordCount)
}

// reads words using strings.Fields and counts them into a map.
func readCountFields(text string) *map[string]uint {
	wordCount := make(map[string]uint)
	for _, word := range strings.Fields(text) {
		wordCount[word]++
	}
	return &wordCount
}

// reads words using DELIMITED regex and counts them into a map.
func readCountRegex(text string) *map[string]uint {
	wordCount := make(map[string]uint)
	for _, word := range DELIMITED.FindAllString(text, ALL) {
		wordCount[strings.ToLower(word)]++
	}
	return &wordCount
}

// reads words using strings.Fields, then applies regex to every word.
func readCountFieldsRegex(text string) *map[string]uint {
	wordCount := make(map[string]uint)
	for _, src := range strings.Fields(text) {
		if word, isWord := transformWord(src); isWord {
			wordCount[word]++
		}
	}
	return &wordCount
}

// reads words using strings.Fields and counts them into a map of words.
func readWordsFields(text string) *map[string]*Word {
	wordCount := make(map[string]*Word)
	for _, word := range strings.Fields(text) {
		cur, ok := wordCount[word]
		if !ok {
			cur = &Word{word, 1, -1}
			wordCount[word] = cur
		} else {
			cur.count++
		}
	}
	return &wordCount
}

// reads words using DELIMITERS regex and counts them into a map.
func readWordsRegex(text string) *map[string]*Word {
	wordCount := make(map[string]*Word)
	for _, match := range DELIMITED.FindAllString(text, ALL) {
		word := strings.ToLower(match)
		cur, ok := wordCount[word]
		if !ok {
			cur = &Word{word, 1, -1}
			wordCount[word] = cur
		} else {
			cur.count++
		}
	}
	return &wordCount
}

// reads words using strings.Fields and counts them into a map of words.
func readWordsFieldsRegex(text string) *map[string]*Word {
	wordCount := make(map[string]*Word)
	for _, src := range strings.Fields(text) {
		if word, isWord := transformWord(src); isWord {
			cur, ok := wordCount[word]
			if !ok {
				cur = &Word{word, 1, -1}
				wordCount[word] = cur
			} else {
				cur.count++
			}
		}
	}
	return &wordCount
}

func transformWord(word string) (string, bool) {
	result := WORD.FindString(word)
	return strings.ToLower(result), len(result) > 0
}

func top10AlgoSimple(wordCount *map[string]uint) []string {
	words := make([]string, 0, len(*wordCount))
	for word := range *wordCount {
		words = append(words, word)
	}
	sort.Slice(words, func(i, j int) bool {
		word1, word2 := words[i], words[j]
		// put the biggest values (most count and lexicographically smaller) first
		return !compareWordsByCount(word1, (*wordCount)[word1], word2, (*wordCount)[word2])
	})
	return words[0:min(len(words), 10)]
}

func top10AlgoArrayHeap(wordCount *map[string]*Word) []string {
	words := make([]*Word, 0, 10)
	for _, cur := range *wordCount {
		words = add(words, cur)
	}
	return createResult(words)
}

func add(words []*Word, word *Word) []*Word {
	index := word.index
	// finds next element which can be beyond the slice
	next := findNext(words, word)
	if len(words) < cap(words) {
		// inserting a new element
		if index == -1 {
			return insert(words, word, next)
		}
	}

	// if new element location needs to be changed
	if next > index+1 {
		// we need to insert dropping first element
		if index == -1 {
			return insert(words, word, next)
		}
		return move(words, word, next)
	}
	return words
}

func insert(words []*Word, word *Word, pos int) []*Word {
	numWords := len(words)
	// if we are inserting and slice will be resized
	if numWords == cap(words) {
		words[0].index = -1
		for _, cur := range words[1:pos] {
			cur.index--
		}
		copy(words, words[1:pos])
		// we deleted one element at the beginning, so index changes
		pos--
		words[pos] = word
		word.index = pos
		return words
	}

	word.index = pos
	// last element
	if pos == numWords {
		return append(words, word)
	}
	for _, cur := range words[pos:numWords] {
		cur.index++
	}
	return append(words[:pos], append([]*Word{word}, words[pos:]...)...)
}

func move(words []*Word, word *Word, pos int) []*Word {
	index := word.index
	next := index + 1
	for _, cur := range words[next:pos] {
		cur.index--
	}
	copy(words[index:], words[next:pos])
	// because we insert up to pos - words[pos] stays intact
	pos--
	word.index = pos
	words[pos] = word
	return words
}

// return index of the first element that is higher than word or the length of the array.
func findNext(words []*Word, word *Word) int {
	// if index is -1 -> we will start from 0
	index := word.index
	start := index + 1
	numWords := len(words)
	if start < numWords {
		for pos, cur := range words[start:numWords] {
			if compareWords(word, cur) {
				return start + pos
			}
		}
	}
	return numWords
}

func createResult(words []*Word) []string {
	numWords := len(words)
	top := make([]string, numWords)
	for index, word := range words {
		// we set from end to the beginning, since last element is numWords - 1, we go backwards
		top[numWords-index-1] = word.word
	}
	return top
}

type WordCount struct {
	word  string
	count uint
}

type MinHeap []WordCount

func (h MinHeap) Len() int { return len(h) }
func (h MinHeap) Less(i, j int) bool {
	return compareWordCount(&h[i], h[j].word, h[j].count)
}

func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *MinHeap) Push(word interface{}) {
	*h = append(*h, word.(WordCount))
}

func (h *MinHeap) Pop() interface{} {
	prev := *h
	numWords := len(prev)
	last := prev[numWords-1]
	*h = prev[0 : numWords-1]
	return last
}

func top10AlgoMinHeap(wordCount *map[string]uint) []string {
	words := make(MinHeap, 0, 10)
	heap.Init(&words)

	for word, count := range *wordCount {
		if words.Len() < cap(words) {
			heap.Push(&words, WordCount{word, count})
			// if new word is higher than the first word in heap
		} else if compareWordCount(&words[0], word, count) {
			heap.Pop(&words)
			heap.Push(&words, WordCount{word, count})
		}
	}

	numWords := words.Len()
	top := make([]string, numWords)
	// only to make sure we get every word
	for index := range words {
		word := heap.Pop(&words).(WordCount)
		// we set from end to the beginning, since last element is numWords - 1, we go backwards
		top[numWords-index-1] = word.word
	}
	return top
}

/*
Finds which of the 2 words should be higher in the words array
First looks into count difference if it is 0, returns lexicographically
earler word
The smaller word either has lower count or is bigger lexicographically.
*/
func compareWordsByCount(word1 string, count1 uint, word2 string, count2 uint) bool {
	if count1 == count2 {
		return word1 > word2
	}
	return count1 < count2
}

/*
Finds which of the 2 words should be higher in the words array
First looks into count difference if it is 0, returns lexicographically
earler word
The smaller word either has lower count or is bigger lexicographically.
*/
func compareWords(word1 *Word, word2 *Word) bool {
	if word1.count == word2.count {
		return word1.word > word2.word
	}
	return word1.count < word2.count
}

/*
Finds which of the 2 words should be higher in the words array
First looks into count difference if it is 0, returns lexicographically
earler word
The smaller word either has lower count or is bigger lexicographically.
*/
func compareWordCount(word *WordCount, word2 string, count2 uint) bool {
	if word.count == count2 {
		return word.word > word2
	}
	return word.count < count2
}
