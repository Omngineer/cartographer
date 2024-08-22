package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"

	svg "github.com/ajstarks/svgo"
)

const BACKGROUND_STYLE = "fill:rgb(227, 203, 168);"
const MIN_GRID_WIDTH = .1
const MIN_ROAD_WIDTH = 1
const ROAD_STYLE_COLOR = "stroke:rgb(77, 42, 24);"

func main() {
	city := City{"Alpha", 0, 0, 0, 0, make(map[P]map[Road]bool), make(map[Road]bool), []float64{}}
	city.addBorderPointAngle(0.0)
	city.addBorderPointAngle(math.Pi / 2)
	city.addBorderPointAngle(math.Pi)
	city.addBorderPointAngle(-math.Pi / 2)
	city.addRoad(P{-100, -100}, P{0, 0})
	city.addRoad(P{0, -10}, P{100, 100})
	city.addRoad(P{0, 0}, P{-50, 0})
	city.finalize()
	fmt.Println(city.summary())
	saveToSvg(city)
	// host(city)
}

type P struct {
	x, y int
}

type Road struct {
	a, b P   // start/end locations
	q    int // quality
}

type City struct {
	name                   string
	maxX, maxY, minX, minY int
	intersections          map[P]map[Road]bool
	roads                  map[Road]bool
	borderPointAngles      []float64 // radians
}

func (c *City) summary() string {
	output := ""
	output += c.summaryOfCity()
	output += c.summaryOfRoads()
	output += c.summaryOfIntersections()
	return output
}

func (c *City) summaryOfCity() string {
	output := "\n"
	output += "City: " + c.name + "\n"
	output += "City Bounds:\n"
	output += fmt.Sprintf("  minX: %d, maxX: %d\n", c.minX, c.maxX)
	output += fmt.Sprintf("  minY: %d, maxY: %d\n", c.minY, c.maxY)

	output += "City Border Points:\n"
	borderPoints := c.findBorderPoints()
	for _, bp := range borderPoints {
		output += fmt.Sprintf("  Location:(%d,%d)\n", bp.x, bp.y)
	}

	return output
}

func (c *City) summaryOfRoads() string {
	output := ""
	output += "Roads:\n"

	roads := make([]Road, len(c.roads))
	i := 0
	for k := range c.roads {
		roads[i] = k
		i++
	}

	for _, road := range roads {
		output += fmt.Sprintf("  s:(%d,%d), e:(%d,%d) q:(%d)\n", road.a.x, road.a.y, road.b.x, road.b.y, road.q)
	}
	return output
}

func (c *City) summaryOfIntersections() string {
	output := ""
	output += "Intersections:\n"

	intersections := make([]P, len(c.intersections))
	i := 0
	for k := range c.intersections {
		intersections[i] = k
		i++
	}

	for _, p := range intersections {
		output += fmt.Sprintf("  i:(%d,%d)\n", p.x, p.y)
	}
	return output
}

func (c *City) finalize() {
	fmt.Println("Finalizing City")

	// Expand the city borders by 10 or 10% whichever is greater to separate the furthest added roads from the border points
	xExpansion := max(10, int(float64(c.maxX-c.minX)/float64(10)))
	yExpansion := max(10, int(float64(c.maxY-c.minY)/float64(10)))
	mapMinX := c.minX - xExpansion
	mapMaxX := c.maxX + xExpansion
	mapMinY := c.minY - yExpansion
	mapMaxY := c.maxY + yExpansion
	c.minX = mapMinX
	c.maxX = mapMaxX
	c.minY = mapMinY
	c.maxY = mapMaxY

	c.attachBorderPoints()
}

func (c *City) getCenter() P {
	return P{(c.minX + c.maxX) / 2, (c.minY + c.maxY) / 2}
}

func (c *City) addRoad(a, b P) {
	c.addRoadWithQuality(a, b, 5)
}

func (c *City) addRoadWithQuality(a, b P, q int) {
	road := Road{a, b, q}
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
	rMinX := min(a.x, b.x)
	if c.minX > rMinX {
		c.minX = rMinX
	}
	rMinY := min(a.y, b.y)
	if c.minY > rMinY {
		c.minY = rMinY
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

func (c *City) addBorderPointAngle(angle float64) {
	adjAngle := math.Mod(angle+(2*math.Pi), 2*math.Pi)
	c.borderPointAngles = append(c.borderPointAngles, adjAngle)
}

func getAngleFromCenter(point, center P) float64 {
	return math.Mod(math.Atan2(float64(center.y), float64(center.x))-math.Atan2(float64(point.y), float64(point.x))+(2*math.Pi), 2*math.Pi)
}

func (c *City) findBorderPoints() []P {
	center := c.getCenter()

	nwAngle := getAngleFromCenter(P{c.minX, c.maxY}, center)
	neAngle := getAngleFromCenter(P{c.maxX, c.maxY}, center)
	seAngle := getAngleFromCenter(P{c.maxX, c.minY}, center)
	swAngle := getAngleFromCenter(P{c.minX, c.minY}, center)

	nBpAngles := []float64{}
	wBpAngles := []float64{}
	sBpAngles := []float64{}
	eBpAngles := []float64{}

	for _, bpAngle := range c.borderPointAngles {
		if bpAngle > nwAngle && bpAngle <= neAngle {
			nBpAngles = append(nBpAngles, bpAngle)
		} else if (bpAngle > neAngle && bpAngle <= (2*math.Pi)) || (bpAngle >= 0 && bpAngle <= seAngle) {
			eBpAngles = append(eBpAngles, bpAngle)
		} else if bpAngle > seAngle && bpAngle <= swAngle {
			sBpAngles = append(sBpAngles, bpAngle)
		} else if bpAngle > swAngle && bpAngle <= nwAngle {
			wBpAngles = append(wBpAngles, bpAngle)
		} else {
			fmt.Printf("NWSE Border not found for: %f\n", bpAngle)
			fmt.Println()
		}
	}

	borderPoints := []P{}
	for _, nAngle := range nBpAngles {
		nBorderFromCenter := c.maxY - center.y
		bpXFromCenter := int(float64(nBorderFromCenter) / math.Tan(nAngle))
		borderPoints = append(borderPoints, P{bpXFromCenter + center.x, c.maxY})
	}
	for _, eAngle := range eBpAngles {
		eBorderFromCenter := c.maxX - center.x
		bpYFromCenter := int(math.Tan(eAngle) / float64(eBorderFromCenter))
		borderPoints = append(borderPoints, P{c.maxX, bpYFromCenter + center.y})
	}
	for _, sAngle := range sBpAngles {
		sBorderFromCenter := c.minY + center.y
		bpXFromCenter := int(float64(sBorderFromCenter) / math.Tan(sAngle))
		borderPoints = append(borderPoints, P{bpXFromCenter + center.x, c.minY})
	}
	for _, wAngle := range wBpAngles {
		wBorderFromCenter := c.minX + center.x
		bpYFromCenter := int(math.Tan(wAngle) / float64(wBorderFromCenter))
		borderPoints = append(borderPoints, P{c.minX, bpYFromCenter + center.y})
	}
	return borderPoints
}

func (c *City) attachBorderPoints() {
	borderPoints := c.findBorderPoints()

	var nearestIntersections []P

	for _, bp := range borderPoints {
		minDist := float64(math.MaxInt32)
		closestIntersection := P{0, 0}
		for i := range c.intersections {
			distToI := math.Sqrt(math.Pow(float64(bp.x-i.x), 2) + math.Pow(float64(bp.y-i.y), 2))
			if distToI < float64(minDist) {
				closestIntersection = i
				minDist = distToI
			}
		}
		nearestIntersections = append(nearestIntersections, closestIntersection)
	}
	for i, bp := range borderPoints {
		c.addRoadWithQuality(nearestIntersections[i], bp, 1)
	}
}

// Cartographer Functions
func saveToSvg(c City) {
	fWriter, err := os.Create(c.name + ".svg")
	if err != nil {
		panic(err)
	}
	defer fWriter.Close()

	s := svg.New(fWriter)
	addCityToSVG(c, s)
	s.End()

	fWriter.Sync()
}

func addCityToSVG(c City, s *svg.SVG) {
	width := c.maxX - c.minX
	height := c.maxY - c.minY
	s.Start(width, height, fmt.Sprintf(`viewBox="0 0 %d %d"`, width, height))
	s.Rect(0, 0, width, height, BACKGROUND_STYLE)
	xCDist := width / 2
	yCDist := height / 2

	// Convert the 0,0 based Points to their SVG pixel location, with 0,0 in the top left
	convertPToSVGP := func(point P) P {
		return P{point.x + xCDist, yCDist - point.y}
	}
	for r := range c.roads {
		rA := convertPToSVGP(r.a)
		rB := convertPToSVGP(r.b)
		s.Polyline([]int{rA.x, rB.x}, []int{rA.y, rB.y}, ROAD_STYLE_COLOR+"stroke-width:"+strconv.Itoa(max(MIN_ROAD_WIDTH, r.q)))
	}
	s.Grid(0, 0, 1, 1, max(width, height), "stroke:gray;stroke-width:"+strconv.FormatFloat(max(float64(MIN_GRID_WIDTH), float64(max(width, height)/1000)), 'E', -1, 64))
}

// Hosts the image at localhost:2003/svg
func host(c City) {
	http.Handle("/svg", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		s := svg.New(w)
		addCityToSVG(c, s)
		s.End()
		fmt.Println(&s)
	}))
	err := http.ListenAndServe(":2003", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
