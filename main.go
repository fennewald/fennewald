package main

import (
    "./gol"
    "./gen"
    "fmt"
)

func main() {
    speed := 0.5 // Speed of simulation in rounds/sec
    process('c', 4, speed)
    process('a', 4, speed)
    process('r', 12, speed)
    process('s', 13, speed)
    process('o', 16, speed)
    process('n', 18, speed)
}

func process(letter rune, rounds uint, speed float64) {
    rleFilename := fmt.Sprintf("./patterns/%c.rle", letter)
    svgFilename := fmt.Sprintf("%c.svg", letter)
    l := gol.FromRLE(rleFilename)
    fmt.Printf("Read %c\n", letter)
    fmt.Println(l.Current.ToString())
    dur := speed * float64(rounds)
    gen.MakeSvg(l, svgFilename, rounds, l.W, l.H, dur)
}

