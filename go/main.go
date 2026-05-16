package main

import (
	"fmt"
	"os"
	"time"

	"github.com/chzyer/readline"
)

func main() {

	rl, err := readline.New("> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	data, err := os.ReadFile("../bigS.txt")
	if err != nil {
		panic("couldn't read bigS: " + err.Error())
	}

	// bigS := "oh 𐀄 shit𐀄s this guy 𐀄hello 𐀄" <-- for testing multi byte runes
	bigS := string(data)
	println(len(bigS))

	loadingBarLen := 51
	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		search := line
		start := time.Now()

		for range loadingBarLen + 8 {
			print("-")
		}
		println()

		loadingBarChunkSize := len(bigS) / loadingBarLen
		lastLoadingBarPrintSize := 0

		fsm := FuzzySearchWith(
			bigS,
			search,
		).CacheDepth(2).MinScore(0.0).TakeProgress(func(progress FuzzySearchProgress) bool {
			if progress.byteOffset-lastLoadingBarPrintSize >= loadingBarChunkSize {
				print("=")
				lastLoadingBarPrintSize = progress.runeOffset
			}

			return false
		}).Run()
		println()

		stop := time.Now()
		elapsed := float64(stop.Sub(start).Microseconds()) / 1000.0

		ss := bigS[fsm.byteOffset : fsm.byteOffset+fsm.byteCount]
		bigerss := bigS[max(0, fsm.byteOffset-200):min(fsm.byteOffset+fsm.byteCount+200, len(bigS)-1)]

		fmt.Println(fsm, ss)
		println()

		fmt.Println(bigerss)
		println(elapsed)
		println("|----------------------------------------------")
	}

}

func use(...any) {
}

// import "strings"
// import "unicode/utf8"

// var s = "abc"

// type BufferedString struct {
// 	s     string
// 	// reader strings.IndexByte()
// 	b1   []rune
// 	b2   []rune
// 	b1i  int
// 	b2i  int
// 	move1 bool
// }

// func NewBufferedString(s string, size int) BufferedString {
// 	new := BufferedString{s: s, b1: make([]rune, 0, size), b2: make([]rune, 0, size)}
// 	return new
// }

// func (bs *BufferedString) charAt(i int) {
// 	[]byte(s)

// 	size := len(bs.b1)
// 	use(size)
// 	// if (strings.NewReader(

// }

// func use(...any) {}

// func (s string) runeAtByte(int byteIndex)

// func runeSize(b byte) int {

// }
