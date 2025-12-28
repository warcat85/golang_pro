package hw03frequencyanalysis

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var text = `Как видите, он  спускается  по  лестнице  вслед  за  своим
	другом   Кристофером   Робином,   головой   вниз,  пересчитывая
	ступеньки собственным затылком:  бум-бум-бум.  Другого  способа
	сходить  с  лестницы  он  пока  не  знает.  Иногда ему, правда,
		кажется, что можно бы найти какой-то другой способ, если бы  он
	только   мог   на  минутку  перестать  бумкать  и  как  следует
	сосредоточиться. Но увы - сосредоточиться-то ему и некогда.
		Как бы то ни было, вот он уже спустился  и  готов  с  вами
	познакомиться.
	- Винни-Пух. Очень приятно!
		Вас,  вероятно,  удивляет, почему его так странно зовут, а
	если вы знаете английский, то вы удивитесь еще больше.
		Это необыкновенное имя подарил ему Кристофер  Робин.  Надо
	вам  сказать,  что  когда-то Кристофер Робин был знаком с одним
	лебедем на пруду, которого он звал Пухом. Для лебедя  это  было
	очень   подходящее  имя,  потому  что  если  ты  зовешь  лебедя
	громко: "Пу-ух! Пу-ух!"- а он  не  откликается,  то  ты  всегда
	можешь  сделать вид, что ты просто понарошку стрелял; а если ты
	звал его тихо, то все подумают, что ты  просто  подул  себе  на
	нос.  Лебедь  потом  куда-то делся, а имя осталось, и Кристофер
	Робин решил отдать его своему медвежонку, чтобы оно не  пропало
	зря.
		А  Винни - так звали самую лучшую, самую добрую медведицу
	в  зоологическом  саду,  которую  очень-очень  любил  Кристофер
	Робин.  А  она  очень-очень  любила  его. Ее ли назвали Винни в
	честь Пуха, или Пуха назвали в ее честь - теперь уже никто  не
	знает,  даже папа Кристофера Робина. Когда-то он знал, а теперь
	забыл.
		Словом, теперь мишку зовут Винни-Пух, и вы знаете почему.
		Иногда Винни-Пух любит вечерком во что-нибудь поиграть,  а
	иногда,  особенно  когда  папа  дома,  он больше любит тихонько
	посидеть у огня и послушать какую-нибудь интересную сказку.
		В этот вечер...`

func TestAlgoSimple(t *testing.T) {
	testAll(t, Top10Simple, false)
}

func TestAlgoSimpleAsterisk(t *testing.T) {
	testAll(t, Top10SimpleAsterisk, true)
}

func TestAlgoArrayHeap(t *testing.T) {
	testAll(t, Top10ArrayHeap, false)
}

func TestAlgoArrayHeapAsterisk(t *testing.T) {
	testAll(t, Top10ArrayHeapAsterisk, true)
}

func TestAlgoPostArrayHeap(t *testing.T) {
	testAll(t, Top10PostArrayHeap, false)
}

func TestAlgoPostArrayHeapAsterisk(t *testing.T) {
	testAll(t, Top10PostArrayHeapAsterisk, true)
}

func TestAlgoPostMinHeap(t *testing.T) {
	testAll(t, Top10PostMinHeap, false)
}

func TestAlgoPostMinHeapAsterisk(t *testing.T) {
	testAll(t, Top10PostMinHeapAsterisk, true)
}

func testAll(t *testing.T, algo func(string) []string, withAsterisk bool) {
	t.Helper()

	testNumWords(t, algo)
	testPunctuation(t, algo, withAsterisk)
	testAlgo(t, algo, withAsterisk)
	testFile(t, algo, withAsterisk)
}

func testAlgo(t *testing.T, algoFunc func(string) []string, withAsterisk bool) {
	t.Helper()

	t.Run("positive test", func(t *testing.T) {
		if withAsterisk {
			expected := []string{
				"а",         // 8
				"он",        // 8
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"в",         // 4
				"его",       // 4
				"если",      // 4
				"кристофер", // 4
				"не",        // 4
			}
			require.Equal(t, expected, algoFunc(text))
		} else {
			expected := []string{
				"он",        // 8
				"а",         // 6
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"-",         // 4
				"Кристофер", // 4
				"если",      // 4
				"не",        // 4
				"то",        // 4
			}
			require.Equal(t, expected, algoFunc(text))
		}
	})

	t.Run("upper lower", func(t *testing.T) {
		var expected []string
		if withAsterisk {
			expected = []string{
				"upper", // 6
			}
		} else {
			expected = []string{
				"Upper", // 3
				"upper", // 3
			}
		}
		require.Equal(t, expected, algoFunc(`
				Upper
				upper
				upper
				upper
				Upper
				Upper
			`))
	})

	t.Run("lower upper", func(t *testing.T) {
		var expected []string
		if withAsterisk {
			expected = []string{
				"upper", // 5
			}
		} else {
			expected = []string{
				"upper", // 3
				"Upper", // 2
			}
		}
		require.Equal(t, expected, algoFunc(`
				Upper
				upper
				upper
				upper
				Upper
			`))
	})

	t.Run("example", func(t *testing.T) {
		var expected []string
		if withAsterisk {
			expected = []string{
				"and",     // 2
				"one",     // 2
				"cat",     // 1
				"cats",    // 1
				"dog",     // 1
				"dog,two", // 1
				"man",     // 1
			}
		} else {
			expected = []string{
				"and",     // 2
				"one",     // 2
				"cat",     // 1
				"cats",    // 1
				"dog,",    // 1
				"dog,two", // 1
				"man",     // 1
			}
		}
		require.Equal(t, expected,
			algoFunc("cat and dog, one dog,two cats and one man"))
	})

	t.Run("random text", func(t *testing.T) {
		var expected []string
		if withAsterisk {
			// omega is lowercase
			expected = []string{
				"◊§∂", // 11
				"∂‡ƒ", // 10
				"∫π¬", // 10
				"≅∇ℓ", // 10
				"∞∝∆", // 9
				"≠µω", // 9
				"ˆ¥≈", // 8
				"ωç∫", // 8
				"∑°∞", // 8
				"∩⊕θ", // 8
			}
		} else {
			expected = []string{
				"◊§∂", // 11
				"∂‡ƒ", // 10
				"∫π¬", // 10
				"≅∇ℓ", // 10
				"∞∝∆", // 9
				"≠µω", // 9
				"ˆ¥≈", // 8
				"Ωç∫", // 8
				"∑°∞", // 8
				"∩⊕θ", // 8
			}
		}
		require.Equal(t, expected,
			algoFunc(
				`
					◊§∂ ≅∇ℓ ∫π¬ ∞∝∆ ◊§∂ ≈√λ ∩⊕θ ≅∇ℓ ≠µω ∑°∞ ∂‡ƒ ˆ¥≈ ≅∇ℓ
					∂‡ƒ ◊§∂ ∑°∞ ∫π¬ ∮κψ ∂‡ƒ Ωç∫ ¬¨˙ ∫π¬ ≠µω ∞∝∆ ◊§∂
					≈√λ ∩⊕θ ˆ¥≈ ∂‡ƒ Ωç∫ ≅∇ℓ ∫π¬ ∞∝∆ ∮κψ ≠µω ≈√λ
					◊§∂ ∑°∞ ∩⊕θ ˆ¥≈ ¬¨˙ ∫π¬ ≅∇ℓ ∂‡ƒ Ωç∫ ∞∝∆ ◊§∂ ≠µω
					≈√λ ∩⊕θ ˆ¥≈ ∮κψ ∑°∞ ≅∇ℓ ∫π¬ ¬¨˙ ∂‡ƒ ◊§∂ Ωç∫
					∞∝∆ ≠µω ≈√λ ∩⊕θ ˆ¥≈ ∮κψ ∑°∞ ≅∇ℓ ∫π¬ ¬¨˙ ∂‡ƒ
					◊§∂ Ωç∫ ∞∝∆ ≠µω ≈√λ ∩⊕θ ˆ¥≈ ∮κψ ∑°∞ ≅∇ℓ ∫π¬
					¬¨˙ ∂‡ƒ ◊§∂ Ωç∫ ∞∝∆ ≠µω ≈√λ ∩⊕θ ˆ¥≈ ∮κψ ∑°∞
					≅∇ℓ ∫π¬ ¬¨˙ ∂‡ƒ ◊§∂ Ωç∫ ∞∝∆ ≠µω ≈√λ ∩⊕θ ˆ¥≈
					∮κψ ∑°∞ ≅∇ℓ ∫π¬ ¬¨˙ ∂‡ƒ ◊§∂ Ωç∫ ∞∝∆ ≠µω
					`))
	})

	t.Run("spaces between punctuation test", func(t *testing.T) {
		var expected []string
		if withAsterisk {
			expected = []string{
				"eighth",       // 1
				"fifth",        // 1
				"first",        // 1
				"ninth",        // 1
				"second",       // 1
				"seventh",      // 1
				"sixth",        // 1
				"tenth",        // 1
				"third,,forth", // 1
			}
		} else {
			expected = []string{
				",ninth,",        // 1
				",second,",       // 1
				",seventh,",      // 1
				",sixth",         // 1
				",third,,forth,", // 1
				"eighth",         // 1
				"fifth",          // 1
				"first,",         // 1
				"tenth",          // 1
			}
		}
		require.Equal(t, expected, algoFunc(
			"   first,   ,second, ,third,,forth"+
				", fifth ,sixth  ,seventh,  eighth   ,ninth,   tenth   "))
	})

	t.Run("punctuation between spaces test", func(t *testing.T) {
		var expected []string
		if withAsterisk {
			// - ignored if more than one between spaces
			expected = []string{
				"-",            // 1
				"eighth",       // 1
				"fifth--sixth", // 1
				"first",        // 1
				"ninth",        // 1
				"second",       // 1
				"seventh",      // 1
				"tenth",        // 1
				"third-forth",  // 1
			}
		} else {
			expected = []string{
				"-",            // 2
				"--",           // 1
				"---",          // 1
				"@first",       // 1
				"eighth",       // 1
				"fifth--sixth", // 1
				"ninth",        // 1
				"second",       // strings.ToLower(word)1
				"seventh",      // 1
				"tenth!",       // 1
			}
		}
		require.Equal(t, expected, algoFunc(
			"@first - second -- third-forth"+
				" fifth--sixth seventh   -   eighth ninth --- tenth!"))
	})

	t.Run("... test", func(t *testing.T) {
		var expected []string
		if withAsterisk {
			// - ignored if more than one between spaces
			expected = []string{
				"fifth...sixth",    // 1
				"first",            // 1
				"forth",            // 1
				"ninth",            // 1
				"second",           // 1
				"seventh...eighth", // 1
				"third",            // 1
			}
		} else {
			expected = []string{
				"...",                    // 1
				"...first",               // 1
				"...forth",               // 1
				"...seventh...eighth...", // 1
				"...third...",            // 1
				"fifth...sixth",          // 1
				"ninth...",               // 1
				"second...",              // 1
			}
		}
		require.Equal(t, expected, algoFunc(
			"...first ... second... ...third... ...forth"+
				" fifth...sixth ...seventh...eighth... ninth..."))
	})
}

func testNumWords(t *testing.T, algoFunc func(string) []string) {
	t.Helper()

	t.Run("no words in empty string", func(t *testing.T) {
		require.Len(t, algoFunc(""), 0)
	})

	t.Run("one word", func(t *testing.T) {
		expected := []string{
			"this-is-one-long-word", // 1
		}
		require.Equal(t, expected, algoFunc("this-is-one-long-word"))
	})
	t.Run("one word five times", func(t *testing.T) {
		expected := []string{
			"this-is-one-long-word", // 5
		}
		require.Equal(t, expected, algoFunc(`
				this-is-one-long-word
				this-is-one-long-word
				this-is-one-long-word
				this-is-one-long-word
				this-is-one-long-word
			`))
	})

	t.Run("nine words", func(t *testing.T) {
		expected := []string{
			"1-one",
			"2-two",
			"3-three",
			"4-four",
			"5-five",
			"6-six",
			"7-seven",
			"8-eight",
			"9-nine",
		}
		require.Equal(t, expected, algoFunc(`
				1-one
				2-two
				3-three
				4-four
				5-five
				6-six
				7-seven
				8-eight
				9-nine
			`))
	})
	t.Run("ten words", func(t *testing.T) {
		expected := []string{
			"01-one",
			"02-two",
			"03-three",
			"04-four",
			"05-five",
			"06-six",
			"07-seven",
			"08-eight",
			"09-nine",
			"10-ten",
		}
		require.Equal(t, expected, algoFunc(`
				01-one
				02-two
				03-three
				04-four
				05-five
				06-six
				07-seven
				08-eight
				09-nine
				10-ten
			`))
	})

	t.Run("twelve words", func(t *testing.T) {
		expected := []string{
			"01-one",
			"02-two",
			"03-three",
			"04-four",
			"05-five",
			"06-six",
			"07-seven",
			"08-eight",
			"09-nine",
			"10-ten",
		}
		require.Equal(t, expected, algoFunc(`
				01-one
				02-two
				03-three
				04-four
				05-five
				06-six
				07-seven
				08-eight
				09-nine
				10-ten
				11-eleven
				12-twelve
			`))
	})
}

func testPunctuation(t *testing.T, algoFunc func(string) []string, withAsterisk bool) {
	t.Helper()

	t.Run("one word with comma", func(t *testing.T) {
		var expected []string
		if withAsterisk {
			expected = []string{
				"this-is-one-long-word", // 1
			}
		} else {
			expected = []string{
				"this-is-one-long-word,", // 1
			}
		}
		require.Equal(t, expected, algoFunc("this-is-one-long-word,"))
	})

	t.Run("one word five times with suffixes", func(t *testing.T) {
		var expected []string
		if withAsterisk {
			expected = []string{
				"this-is-one-long-word", // 5
			}
		} else {
			expected = []string{
				"this-is-one-long-word!", // 1
				"this-is-one-long-word,", // 1
				"this-is-one-long-word-", // 1
				"this-is-one-long-word.", // 1
				"this-is-one-long-word:", // 1
			}
		}
		require.Equal(t, expected, algoFunc(`
				this-is-one-long-word,
				this-is-one-long-word:
				this-is-one-long-word-
				this-is-one-long-word!
				this-is-one-long-word.
			`))
	})

	t.Run("one word five times with delimiters", func(t *testing.T) {
		var expected []string
		if withAsterisk {
			// $ and ^ are not punctuation characters
			expected = []string{
				"this-is-one-long-word",  // 3
				"$this-is-one-long-word", // 1
				"^this-is-one-long-word", // 1
			}
		} else {
			expected = []string{
				"#this-is-one-long-word:", // 1
				"$this-is-one-long-word-", // 1
				"%this-is-one-long-word!", // 1
				"@this-is-one-long-word,", // 1
				"^this-is-one-long-word.", // 1
			}
		}
		require.Equal(t, expected, algoFunc(`
				@this-is-one-long-word,
				#this-is-one-long-word:
				$this-is-one-long-word-
				%this-is-one-long-word!
				^this-is-one-long-word.
			`))
	})

	t.Run("spaces between punctuation test", func(t *testing.T) {
		var expected []string
		if withAsterisk {
			expected = []string{
				"eighth",       // 1
				"fifth",        // 1
				"first",        // 1
				"ninth",        // 1
				"second",       // 1
				"seventh",      // 1
				"sixth",        // 1
				"tenth",        // 1
				"third,,forth", // 1
			}
		} else {
			expected = []string{
				",ninth,",        // 1
				",second,",       // 1
				",seventh,",      // 1
				",sixth",         // 1
				",third,,forth,", // 1
				"eighth",         // 1
				"fifth",          // 1
				"first,",         // 1
				"tenth",          // 1
			}
		}
		require.Equal(t, expected, algoFunc(
			"   first,   ,second, ,third,,forth"+
				", fifth ,sixth  ,seventh,  eighth   ,ninth,   tenth   "))
	})

	t.Run("punctuation between spaces test", func(t *testing.T) {
		var expected []string
		if withAsterisk {
			// - ignored if more than one between spaces
			expected = []string{
				"-",            // 1
				"eighth",       // 1
				"fifth--sixth", // 1
				"first",        // 1
				"ninth",        // 1
				"second",       // 1
				"seventh",      // 1
				"tenth",        // 1
				"third-forth",  // 1
			}
		} else {
			expected = []string{
				"-",            // 2
				"--",           // 1
				"---",          // 1
				"@first",       // 1
				"eighth",       // 1
				"fifth--sixth", // 1
				"ninth",        // 1
				"second",       // strings.ToLower(word)1
				"seventh",      // 1
				"tenth!",       // 1
			}
		}
		require.Equal(t, expected, algoFunc(
			"@first - second -- third-forth"+
				" fifth--sixth seventh   -   eighth ninth --- tenth!"))
	})

	t.Run("... test", func(t *testing.T) {
		var expected []string
		if withAsterisk {
			// - ignored if more than one between spaces
			expected = []string{
				"fifth...sixth",    // 1
				"first",            // 1
				"forth",            // 1
				"ninth",            // 1
				"second",           // 1
				"seventh...eighth", // 1
				"third",            // 1
			}
		} else {
			expected = []string{
				"...",                    // 1
				"...first",               // 1
				"...forth",               // 1
				"...seventh...eighth...", // 1
				"...third...",            // 1
				"fifth...sixth",          // 1
				"ninth...",               // 1
				"second...",              // 1
			}
		}
		require.Equal(t, expected, algoFunc(
			"...first ... second... ...third... ...forth"+
				" fifth...sixth ...seventh...eighth... ninth..."))
	})
}

func readFile(tb testing.TB) string {
	tb.Helper()
	in, err := os.ReadFile("war_and_peace.txt")
	if err != nil {
		tb.Fatal(err)
	}
	return string(in)
}

func testFile(t *testing.T, algoFunc func(string) []string, withAsterisk bool) {
	t.Helper()

	t.Run("file", func(t *testing.T) {
		var expected []string
		if withAsterisk {
			expected = []string{
				"и",   // 10643
				"в",   // 5307
				"не",  // 4402
				"что", // 3939
				"он",  // 3818
				"на",  // 3383
				"с",   // 3103
				"как", // 2149
				"я",   // 1949
				"его", // 1927
			}
		} else {
			expected = []string{
				"и",   // 9738
				"–",   // 8186
				"в",   // 4846
				"не",  // 4215
				"что", // 3474
				"на",  // 3192
				"с",   // 3009
				"он",  // 2423
				"как", // 1856
				"к",   // 1770
			}
		}
		str := readFile(t)
		require.Equal(t, expected,
			algoFunc(str))
	})
}

func benchmarkFile(b *testing.B, algo func(string) []string) {
	b.Helper()

	str := readFile(b)
	b.ResetTimer()
	for range b.N {
		algo(str)
	}
}

func benchmarkAlgo(
	b *testing.B, reader func(string) *map[string]uint, algo func(*map[string]uint) []string,
) {
	b.Helper()
	str := readFile(b)
	wordCount := reader(str)
	b.ResetTimer()
	for range b.N {
		algo(wordCount)
	}
}

func benchmarkAlgoArrayHeap(
	b *testing.B, reader func(string) *map[string]*Word, algo func(*map[string]*Word) []string,
) {
	b.Helper()
	str := readFile(b)
	wordCount := reader(str)
	b.ResetTimer()
	for range b.N {
		algo(wordCount)
	}
}

func BenchmarkFileSimpleFields(b *testing.B) {
	benchmarkFile(b, Top10Simple)
}

func BenchmarkFileArrayHeapFields(b *testing.B) {
	benchmarkFile(b, Top10ArrayHeap)
}

func BenchmarkFilePostArrayHeapFields(b *testing.B) {
	benchmarkFile(b, Top10PostArrayHeap)
}

func BenchmarkFilePostMinHeapFields(b *testing.B) {
	benchmarkFile(b, Top10PostMinHeap)
}

func BenchmarkFileSimpleRegex(b *testing.B) {
	benchmarkFile(b, Top10SimpleAsterisk)
}

func BenchmarkFileArrayHeapRegex(b *testing.B) {
	benchmarkFile(b, Top10ArrayHeapAsterisk)
}

func BenchmarkFilePostArrayHeapRegex(b *testing.B) {
	benchmarkFile(b, Top10PostArrayHeapAsterisk)
}

func BenchmarkFilePostMinHeapRegex(b *testing.B) {
	benchmarkFile(b, Top10PostMinHeapAsterisk)
}

func BenchmarkAlgoSimpleFields(b *testing.B) {
	benchmarkAlgo(b, readCountFields, top10AlgoSimple)
}

func BenchmarkAlgoSimpleRegex(b *testing.B) {
	benchmarkAlgo(b, readCountRegex, top10AlgoSimple)
}

func BenchmarkAlgoPostArrayHeapFields(b *testing.B) {
	benchmarkAlgoArrayHeap(b, readWordsFields, top10AlgoArrayHeap)
}

func BenchmarkAlgoPostArrayHeapRegex(b *testing.B) {
	benchmarkAlgoArrayHeap(b, readWordsRegex, top10AlgoArrayHeap)
}

func BenchmarkAlgoPostMinHeapFields(b *testing.B) {
	benchmarkAlgo(b, readCountFields, top10AlgoMinHeap)
}

func BenchmarkAlgoPostMinHeapRegex(b *testing.B) {
	benchmarkAlgo(b, readCountRegex, top10AlgoMinHeap)
}
