package crossword

import (
	"bytes"
	"math"
	"math/rand"
	"regexp"
	"slices"
	"strings"
)

type Crossword struct {
	Height int
	Width  int
	Cells  [][]byte
}

type CrosswordEntry struct {
	Row        int
	Column     int
	Horizontal bool
	Word       string
}

type Target struct {
	Row    int
	Column int
	Across bool
}

type State struct {
	cbytes []byte
	target Target
	picked string
	kaput  []string
}

type WordList []string

func New(height, width int) *Crossword {
	c := Crossword{Height: height, Width: width}
	c.Cells = make([][]byte, height)
	for i := 0; i < height; i++ {
		c.Cells[i] = make([]byte, width)
	}
	for i := range c.Height {
		for j := range c.Width {
			c.Cells[i][j] = 46
		}
	}
	return &c
}

func (c *Crossword) SetCell(x, y int, b byte) {
	c.Cells[y][x] = b
}

func (c *Crossword) GetCell(x, y int) byte {
	return c.Cells[y][x]
}

func (c *Crossword) GetWord(x, y int, across bool) string {
	if across {
		return c.getWordAcross(x, y)
	}
	return c.getWordDown(x, y)
}

func (c *Crossword) getWordAcross(row, col int) string {
	b := make([]byte, c.Width)

	i := col
	for i < c.Width && c.Cells[row][i] != 45 {
		b[i] = c.Cells[row][i]
		i++
	}

	return string(b[col:i])
}

func (c *Crossword) getWordDown(row, col int) string {
	b := make([]byte, c.Height)

	i := row
	for i < c.Height && c.Cells[i][col] != 45 {
		b[i] = c.Cells[i][col]
		i++
	}

	return string(b[row:i])
}

func (c *Crossword) SetWord(x, y int, across bool, word []byte) {
	if across {
		c.setWordAcross(x, y, word)
	} else {
		c.setWordDown(x, y, word)
	}
}

func (c *Crossword) setWordAcross(row, col int, word []byte) {
	l := min(len(word), c.Width-col)

	for i := range l {
		c.Cells[row][i+col] = word[i]
	}
}

func (c *Crossword) setWordDown(row, col int, word []byte) {
	l := min(len(word), c.Height-row)

	for i := range l {
		c.Cells[i+row][col] = word[i]
	}
}

func (c *Crossword) ToString() string {
	b := make([]string, c.Height+2)

	b[0] = strings.Join([]string{"+", "-----", "+"}, "")
	for i := range c.Height {
		b[i+1] = "|" + string(c.Cells[i]) + "|"
	}
	b[c.Height+1] = strings.Join([]string{"+", "-----", "+"}, "")
	return strings.Join(b, "\n")
}

func (c *Crossword) FromString(s string) {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		for j, r := range line {
			c.Cells[i][j] = byte(r)
		}
	}
}

func (c *Crossword) ToByteArray() []byte {
	return bytes.Join(c.Cells, nil)
}

func (c *Crossword) FromByteArray(b []byte) {
	for row := range c.Height {
		copy(c.Cells[row], b[row*c.Width:(row+1)*c.Width])
	}
}

func (c *Crossword) GetEntries() []CrosswordEntry {
	entries := make([]CrosswordEntry, 0)

	for row := range c.Height {
		for col := range c.Width {
			if c.Cells[row][col] != 45 {
				if col == 0 || c.Cells[row][col-1] == 45 {
					word := c.getWordAcross(row, col)
					if len(word) > 1 {
						entries = append(entries, CrosswordEntry{Row: row, Column: col, Horizontal: true, Word: word})
					} else {
						// c.SetCell(col, row, 45)
					}
				}
				if row == 0 || c.Cells[row-1][col] == 45 {
					word := c.getWordDown(row, col)
					if len(word) > 1 {
						entries = append(entries, CrosswordEntry{Row: row, Column: col, Horizontal: false, Word: word})
					} else {
						// c.SetCell(col, row, 45)
					}
				}
			}
		}
	}

	return entries
}

func (c *Crossword) Fill(wordList *WordList) bool {
	sa := []State{}
	sa = append(sa, State{c.ToByteArray(), Target{2, 0, true}, "", []string{}})

	var done = false
	for iterations := 0; !done && iterations < 1000; iterations++ {
		s := &sa[len(sa)-1]
		entries := c.GetEntries()
		numMatches := make([]int, len(entries))
		for i := range numMatches {
			numMatches[i] = math.MaxInt
		}
		minMatch := math.MaxInt
		minMatchIndex := -1

		for i, entry := range entries {
			if strings.Contains(entry.Word, ".") {
				re := regexp.MustCompile(entry.Word)
				numMatches[i] = len(re.FindAllStringIndex((*wordList)[len(entry.Word)], -1))
				if numMatches[i] < minMatch {
					minMatch = numMatches[i]
					minMatchIndex = i
				}
			}
		}

		if minMatchIndex == -1 {
			correct := true
			for _, entry := range entries {
				re := regexp.MustCompile(entry.Word)
				found := re.FindAllString((*wordList)[len(entry.Word)], -1)
				if len(found) == 0 {
					correct = false
					break
				}
			}

			if correct {
				done = true
				break
			} else {
				sa = sa[:len(sa)-1]
				s = &sa[len(sa)-1]
				s.kaput = append(s.kaput, s.picked)
				c.FromByteArray(s.cbytes)
				continue
			}
		}

		entry := entries[minMatchIndex]

		matches := regexp.MustCompile(entry.Word).FindAllString((*wordList)[len(entry.Word)], -1)

		usable := slices.DeleteFunc(matches, func(match string) bool {
			return slices.Contains(s.kaput, match)
		})

		if len(usable) == 0 {
			sa = sa[:len(sa)-1]
			s = &sa[len(sa)-1]
			s.kaput = append(s.kaput, s.picked)
			c.FromByteArray(s.cbytes)
			continue
		}

		s.picked = usable[rand.Intn(len(usable))]
		s.target = Target{entry.Row, entry.Column, entry.Horizontal}
		c.SetWord(entry.Row, entry.Column, entry.Horizontal, []byte(s.picked))

		newState := State{c.ToByteArray(), Target{0, 0, true}, "", []string{}}
		sa = append(sa, newState)
	}

	return done
}
