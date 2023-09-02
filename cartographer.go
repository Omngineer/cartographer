package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	svg "github.com/ajstarks/svgo"
)

type P struct {
	x, y int
}

type Road struct {
	a, b P
}

type City struct {
	name          string
	maxX, maxY    int
	intersections map[P]map[Road]bool
	roads         map[Road]bool
}

func (c *City) addRoad(a, b P) {
	road := Road{a, b}
	_, ok := c.roads[road]
	if ok {
		return
	}

	rMaxX := max(a.x, b.x)
	if c.maxX < rMaxX {
		c.maxX = rMaxX
	}
	rMaxY := max(a.y, b.y)
	if c.maxY < rMaxY {
		c.maxY = rMaxY
	}

	appendToIntersection := func(p P, r Road) {
		if intAtP, ok := c.intersections[p]; ok {
			_, roadAlreadyExists := intAtP[r]
			if !roadAlreadyExists {
				intAtP[r] = true
			}
		} else {
			newInt := make(map[Road]bool)
			newInt[r] = true
			c.intersections[p] = newInt
		}
	}

	appendToIntersection(a, road)
	appendToIntersection(b, road)
	c.roads[road] = true
}

func main() {
	city := City{"alpha", 1, 1, make(map[P]map[Road]bool), make(map[Road]bool)}
	city.addRoad(P{0, 0}, P{10, 10})
	city.addRoad(P{2, 7}, P{1, 17})

	fmt.Println(city)
	host(city)
}

func buildCitySVG(c City, s *svg.SVG) {
	s.Start(c.maxX, c.maxY)
	s.Rect(0, 0, c.maxX, c.maxY, "fill:rgb(227, 203, 168);")
	rWidth := max(2, max(c.maxX, c.maxY)/100)
	for r := range c.roads {
		s.Polyline([]int{r.a.x, r.b.x}, []int{r.a.y, r.b.y}, "stroke:rgb(77, 42, 24);stroke-width:"+strconv.Itoa(rWidth))
	}
	s.End()
}

// Hosts the image at localhost:2003/svg
func host(c City) {
	http.Handle("/svg", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		s := svg.New(w)
		buildCitySVG(c, s)
	}))
	err := http.ListenAndServe(":2003", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
