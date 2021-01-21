package gen

import (
	"../bitstring"
	"../gol"
	"fmt"
	"os"
	"strings"
)

func MakeSvg(l *gol.Life, filename string, rounds, width, height uint, dur float64) {
	// Generate frame data
	fmt.Println("Generating states")
	cells := generateStates(l, rounds)
	// Open file for write
	fmt.Println("Writing file")
	f, _ := os.Create(filename)
	defer f.Close()
	// Write initial junk
	f.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"no\"?>")
	f.WriteString("<!DOCTYPE svg>")
	f.WriteString("<svg xmlns=\"http://www.w3.org/2000/svg\" ")
	f.WriteString(fmt.Sprintf("viewbox=\"0 0 %d %d\">", l.W, l.H))
	//f.WriteString(fmt.Sprintf("width=\"%dpx\" height=\"%dpx\">", width, height))
	// Write every cell
	for _, cell := range cells {
		f.WriteString(cell.toSvg(dur, rounds))
	}
	// Write closing
	f.WriteString("</svg>")
}

// Struct to hold a state
type cellState struct {
	x, y  uint
	start uint
	dur   uint
}

// Struct to hold cell information
type cell struct {
	states []cellState
}

func generateStates(l *gol.Life, rounds uint) (cells []cell) {
	// Variables
	var i uint
	var id int
	var plot [][]int
	var open, cooldown []int
	var born, died bitstring.Bitstring

	// Variable init
	plot = make([][]int, l.W)
	for i = 0; i < l.W; i++ {
		plot[i] = make([]int, l.H)
	}

	// Handle starting
	for coord := range l.Current.List() {
		plot[coord[0]][coord[1]] = len(cells)
		cells = append(cells, *newCell(coord[0], coord[1], 0))
	}

	// Calculate rounds
	for i = 1; i < rounds; i++ {
		// Pull cells off cooldown
		if len(cooldown) > 0 {
			open = append(open, cooldown...)
			cooldown = nil
		}
		// Step
		l.Step()

		// Handle deaths
		died = *l.Past.NowOn(l.Current)
		for coord := range died.List() {
			id = plot[coord[0]][coord[1]]
			plot[coord[0]][coord[1]] = -1
			cells[id].setLastDuration(i)
			cooldown = append(cooldown, id)
		}

		// Handle births
		born = *l.Current.NowOn(l.Past)
		for coord := range born.List() {
			if len(open) > 0 {
				id = open[0]
				plot[coord[0]][coord[1]] = id
				open = open[1:]
				cells[id].addState(coord[0], coord[1], i)
			} else {
				id = len(cells)
				plot[coord[0]][coord[1]] = id
				cells = append(cells, *newCell(coord[0], coord[1], i))
			}
		}
	}

	// Set duration for survivors
	for coord := range l.Current.List() {
		id = plot[coord[0]][coord[1]]
		cells[id].setLastDuration(rounds)
	}
	return
}

// Methods for cells
func newCell(x, y, time uint) *cell {
	c := cell{}
	c.addState(x, y, time)
	return &c
}

func (c *cell) setLastDuration(time uint) {
	index := len(c.states) - 1
	c.states[index].dur = time - c.states[index].start
}

func (c *cell) addState(x, y, time uint) {
	c.states = append(c.states, cellState{x: x, y: y, start: time})
}

func (c *cell) coords(rounds int) []bool {
	status := make([]bool, rounds)
	for i := 0; i < rounds; i++ {
		status[i] = false
	}
	for _, state := range c.states {
		for i := state.start; i < (state.dur + state.start); i++ {
			status[i] = true
		}
	}
	return status
}

func (c *cell) iterateStates(rounds uint) (chan uint, chan uint, chan uint) {
    chOpacity := make(chan uint, rounds)
    chX := make(chan uint, rounds)
    chY := make(chan uint, rounds)
    go func() {
        var stateIndex, i uint
        state := c.states[0]
        stateIndex = 0
        for i = 0; i < rounds; i++ {
            if i < state.start {
                // Before start
                chOpacity <- 1
                chX <- state.x
                chY <- state.y
            } else if i < (state.start + state.dur) {
                // Inside of state
                chOpacity <- 1
                chX <- state.x
                chY <- state.y
            } else {
                // Done
                chOpacity <- 0
                chX <- state.x
                chY <- state.y
            }

            if i == (state.start + state.dur - 1) {
                stateIndex ++
                if stateIndex >= uint(len(c.states)) {
                    state = c.states[0]
                } else {
                    state = c.states[stateIndex]
                }
            }
        }
        close(chOpacity)
        close(chX)
        close(chY)
    }()
    return chOpacity, chX, chY
}

func (c *cell) toSvg(dur float64, rounds uint) string {
    chOpacity, chX, chY := c.iterateStates(rounds)
	// Animate opacity
	opacityStr := animateStr(chOpacity, "opacity", dur)
	// Animate x
	xStr := animateStrTransition(chX, "x", dur)
	// Animate y
	yStr := animateStrTransition(chY, "y", dur)
	return fmt.Sprintf(
		"<rect width=\"1\" height=\"1\" fill=\"#000\" opacity=\"1\" x=\"%d\" y=\"%d\">%s%s%s</rect>",
		c.states[0].x,
		c.states[0].y,
		opacityStr,
		xStr,
		yStr)
}

func animateStr(ch chan uint, name string, dur float64) string {
	var strArr []string
	var last uint
	unique := false
	first := true
	for val := range ch {
		if first {
			last = val
			first = false
		} else if val != last {
			unique = true
		}
		strArr = append(strArr, fmt.Sprintf("%v;%v", val, val))
	}
	if !unique {
		return ""
	}
	return fmt.Sprintf(
		"<animate attributeName=\"%s\" values=\"%s\" dur=\"%fs\" repeatCount=\"indefinite\"/>",
		name, strings.Join(strArr, ";"), dur)
}

func animateStrTransition(ch chan uint, name string, dur float64) string {
	var strArr []string
	var last uint
	unique := false
	first := true
	for val := range ch {
		if first {
			last = val
			first = false
		} else if val != last {
			unique = true
			strArr = append(strArr, fmt.Sprintf("%v;%v", last, val))
		} else {
			strArr = append(strArr, fmt.Sprintf("%v;%v", val, val))
		}
		last = val
	}
	if !unique {
		return ""
	}
	return fmt.Sprintf(
		"<animate attributeName=\"%s\" values=\"%s\" dur=\"%fs\" repeatCount=\"indefinite\"/>",
		name, strings.Join(strArr, ";"), dur)
}
