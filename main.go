package main

import (
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	_ "image/png"
)

type Imageinfo struct {
	XMLName xml.Name  `xml:"imageinfo"`
	CellsX  int       `xml:"cells_x"`
	CellsY  int       `xml:"cells_y"`
	Segment []Segment `xml:"segment"`
}

func (i *Imageinfo) String() string {
	lines := []string{}
	lines = append(lines, strings.Repeat("-", i.CellsX*2+2))
	for _, v := range i.Segment {
		lines = append(lines, v.String())
	}
	lines = append(lines, strings.Repeat("-", i.CellsX*2+2))
	return strings.Join(lines, "\n")
}

type Segment struct {
	X1    int `xml:"x1,attr"`
	X2    int `xml:"x2,attr"`
	Y1    int `xml:"y1,attr"`
	Y2    int `xml:"y2,attr"`
	maxwd int
}

func (s *Segment) String() string {
	lines := []string{}
	line := []rune{}
	line = append(line, '|')
	for y := s.Y1 - 1; y < s.Y2; y++ {
		for x := 1; x < s.X1; x++ {
			line = append(line, ' ')
			line = append(line, ' ')
		}
		for x := s.X1; x <= s.X2; x++ {
			line = append(line, '■')
			line = append(line, '■')
		}
		for x := s.X2; x < s.maxwd; x++ {
			line = append(line, ' ')
			line = append(line, ' ')
		}
		line = append(line, '|')
		lines = append(lines, string(line))
	}
	return strings.Join(lines, "\n")
}

func isBlack(col color.Color) bool {
	switch x := col.(type) {
	case color.Gray:
		return x.Y < 0xfa
	default:
		log.Fatalf("Unknown type for color %#v", col)
	}

	return true
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println(os.Args[0], "<inputfile>")
		os.Exit(0)
	}
	imagepath := os.Args[1]
	name := strings.TrimSuffix(imagepath, filepath.Ext(imagepath))
	resolution := 40

	// Run imagemagick to get a max(40x40) pixel image
	cmd := exec.Command("convert", imagepath, "-background", "white", "-alpha", "remove", "-type", "Grayscale", "-resize", fmt.Sprintf("%dx%d", resolution, resolution), "-threshold", "97%", "png:-")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	img, _, err := image.Decode(stdout)
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}

	// Now construct the XML
	bounds := img.Bounds()
	ii := Imageinfo{}
	ii.CellsX = bounds.Dx()
	ii.CellsY = bounds.Dy()
	for y := 0; y < bounds.Dy(); y++ {
		s := Segment{maxwd: ii.CellsX}
		hasBlackPixel := false
		for x := 0; x < bounds.Dx(); x++ {
			s.Y1 = y + 1
			s.Y2 = y + 1
			if isBlack(img.At(x, y)) {
				hasBlackPixel = true
				if s.X1 == 0 {
					s.X1 = x + 1
				}
				s.X2 = x + 1
			}
		}
		if hasBlackPixel {
			ii.Segment = append(ii.Segment, s)
		}
	}
	b, err := xml.MarshalIndent(ii, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ii.String())

	f, err := os.OpenFile(name+".xml", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	_, err = f.Write(b)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	f.Close()
}
