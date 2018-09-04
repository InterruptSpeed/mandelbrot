package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"runtime/trace"
	"sync"
)

const (
	output = "out.png"
	width  = 2048
	height = 2048
)

func main() {
	trace.Start(os.Stdout)
	defer trace.Stop()

	f, err := os.Create(output)
	if err != nil {
		log.Fatal(err)
	}

	img := createRowWorkers(width, height)

	if err = png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}

func createSeq(width, height int) image.Image {
	m := image.NewGray(image.Rect(0, 0, width, height))
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			m.Set(i, j, pixel(i, j, width, height))
		}
	}
	return m
}

func createPixel(width, height int) image.Image {
	m := image.NewGray(image.Rect(0, 0, width, height))
	var w sync.WaitGroup
	w.Add(width * height)
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			go func(i, j int) {
				m.Set(i, j, pixel(i, j, width, height))
				w.Done()
			}(i, j)
		}
	}
	w.Wait()
	return m
}

func createRow(width, height int) image.Image {
	m := image.NewGray(image.Rect(0, 0, width, height))
	var w sync.WaitGroup
	w.Add(width)
	for i := 0; i < width; i++ {
		go func(i int) {
			for j := 0; j < height; j++ {
				m.Set(i, j, pixel(i, j, width, height))
			}
			w.Done()
		}(i)
	}
	w.Wait()
	return m
}

func createWorkers(width, height int) image.Image {
	m := image.NewGray(image.Rect(0, 0, width, height))

	type px struct{ x, y int }

	c := make(chan px, width*height)

	var w sync.WaitGroup
	w.Add(8)
	for i := 0; i < 8; i++ {
		go func() {
			for px := range c {
				m.Set(px.x, px.y, pixel(px.x, px.y, width, height))
			}
			w.Done()
		}()
	}
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			c <- px{i, j}
		}
	}
	close(c)
	w.Wait()

	return m
}

func createRowWorkers(width, height int) image.Image {
	m := image.NewGray(image.Rect(0, 0, width, height))

	c := make(chan int, width)

	var w sync.WaitGroup
	w.Add(8)
	for i := 0; i < 8; i++ {
		go func() {
			for i := range c {
				for j := 0; j < height; j++ {
					m.Set(i, j, pixel(i, j, width, height))
				}
			}
			w.Done()
		}()
	}
	for i := 0; i < width; i++ {
		c <- i
	}
	close(c)
	w.Wait()

	return m
}

func pixel(i, j, width, height int) color.Color {
	// Play with this constant to increase the complexity of the fractal.
	// In the justforfunc.com video this was set to 4.
	const complexity = 1024
	xi := norm(i, width, -1.0, 2)
	yi := norm(j, height, -1, 1)
	const maxI = 1000
	x, y := 0., 0.
	for i := 0; (x*x+y*y < complexity) && i < maxI; i++ {
		x, y = x*x-y*y+xi, 2*x*y+yi
	}
	return color.Gray{uint8(x)}
}

func norm(x, total int, min, max float64) float64 {
	return (max-min)*float64(x)/float64(total) - max
}
