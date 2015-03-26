package main

import "math"

import "fmt"

import ()

func main() {

	deltaY := int(114) - int(117)
	deltaX := int(142) - int(93)

	angle := -math.Atan2(float64(deltaY), float64(deltaX)) * 180 / math.Pi

	fmt.Println("angle ", angle)
}
