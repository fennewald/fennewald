package gol

import (
	"../bitstring"
	"sync"
	"os"
	"bufio"
	"fmt"
)

type Life struct {
	W, H          uint
	a, b          bitstring.Bitstring
	Current, Past *bitstring.Bitstring
}

func NewLife(w, h uint) *Life {
	l := Life{
		W: w,
		H: h,
	}
	l.a = *bitstring.NewBitstring(w, h)
	l.b = *bitstring.NewBitstring(w, h)
	l.Current = &l.a
	l.Past = &l.b
	return &l
}

func FromRLE(filename string) *Life {
	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(f)
	// Read comment lines at start
	for scanner.Scan(); scanner.Text()[0] == '#'; scanner.Scan() {
		fmt.Printf("Comment at head of file %s: %s\n", filename, scanner.Text())
	}
	// Read dimensions
	var w, h uint
	fmt.Sscanf(scanner.Text(), "x = %d, y = %d", &w, &h)
	fmt.Printf("Read filename %s width: %d, height %d\n", filename, w, h)
	l := NewLife(w, h)
	// Read content
	var i, j, coeffecient uint
	i = 0
	coeffecient = 0
	for scanner.Scan() {
		for _, c := range scanner.Text() {
			switch c {
			case '0':
				coeffecient *= 10
			case '1':
				coeffecient *= 10
				coeffecient += 1
			case '2':
				coeffecient *= 10
				coeffecient += 2
			case '3':
				coeffecient *= 10
				coeffecient += 3
			case '4':
				coeffecient *= 10
				coeffecient += 4
			case '5':
				coeffecient *= 10
				coeffecient += 5
			case '6':
				coeffecient *= 10
				coeffecient += 6
			case '7':
				coeffecient *= 10
				coeffecient += 7
			case '8':
				coeffecient *= 10
				coeffecient += 8
			case '9':
				coeffecient *= 10
				coeffecient += 9
			case 'b':
    			if coeffecient == 0 {
        			i ++
    			} else {
    				i += coeffecient
    				coeffecient = 0
    			}
			case 'o':
    			if coeffecient == 0 {
        			coeffecient = 1
    			}
				for j = 0; j < coeffecient; j++ {
					l.Current.Set(i, 0, true)
					i++
				}
				coeffecient = 0
				// TODO: Make this go block by block
			case '$':
    			if coeffecient == 0 {
        			coeffecient = 1
    			}
				if i%w != 0 {
    				i += w - (i % w)
				}
				coeffecient --
				i += (w * coeffecient)
				coeffecient = 0
			case '!':
				if coeffecient != 0 {
					fmt.Printf("Error, ended rle with coeffecient of %d\n", coeffecient)
				}
            	return l
			}
		}
	}
	return l
}

func (l *Life) StepCell(x, y uint, wg *sync.WaitGroup) {
	sum := 0
	if l.Past.Get(x-1, y-1) {
		sum++
	}
	if l.Past.Get(x-1, y) {
		sum++
	}
	if l.Past.Get(x-1, y+1) {
		sum++
	}
	if l.Past.Get(x, y-1) {
		sum++
	}
	if l.Past.Get(x, y+1) {
		sum++
	}
	if l.Past.Get(x+1, y-1) {
		sum++
	}
	if l.Past.Get(x+1, y) {
		sum++
	}
	if l.Past.Get(x+1, y+1) {
		sum++
	}
	if sum == 3 || (sum == 2 && l.Past.Get(x, y)) {
		l.Current.Set(x, y, true)
	} else {
		l.Current.Set(x, y, false)
	}
	wg.Done()
}

func (l *Life) SetPattern(x, y uint, str string) {
	start := x
	for _, c := range str {
		if c == '\n' {
			x = start
			y++
		} else {
			if c == '.' {
				l.Current.Set(x, y, false)
			} else {
				l.Current.Set(x, y, true)
			}
			x++
		}
	}
}

func (l *Life) Step() {
	var wg sync.WaitGroup
	var x, y uint
	l.Current, l.Past = l.Past, l.Current
	for x = 0; x < l.W; x++ {
		for y = 0; y < l.H; y++ {
			wg.Add(1)
			go l.StepCell(x, y, &wg)
		}
	}
	wg.Wait()
}

func (l *Life) ToString() string {
	return l.Current.ToString()
}
