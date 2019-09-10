package main

//go:generate go get github.com/rakyll/statik
//go:generate statik

import (
	"bytes"
	"flag"
	"io/ioutil"

	"image"
	"image/draw"
	"image/png"
	"log"
	"net/http"
	"os"

	"github.com/disintegration/imaging"
	"github.com/mattn/go-sixel"
	_ "github.com/calorie/longcat/statik"
	"github.com/rakyll/statik/fs"
)

func loadImage(fs http.FileSystem, n string) (image.Image, error) {
	f, err := fs.Open(n)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	return png.Decode(f)
}

func saveImage(filename string, img image.Image) error {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, buf.Bytes(), 0644)
}

func main() {
	var nlong int
	var nrows int
	var rinterval float64
	var flipH bool
	var flipV bool
	var filename string

	flag.IntVar(&nlong, "n", 1, "how long cat")
	flag.IntVar(&nrows, "l", 1, "number of rows")
	flag.Float64Var(&rinterval, "i", 1.0, "rate of intervals")
	flag.BoolVar(&flipH, "r", false, "flip holizontal")
	flag.BoolVar(&flipV, "R", false, "flip vertical")
	flag.StringVar(&filename, "o", "", "output image file")
	flag.Parse()

	fs, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	img1, _ := loadImage(fs, "/data01.png")
	img2, _ := loadImage(fs, "/data02.png")

	if flipH {
		img1 = imaging.FlipH(img1)
		img2 = imaging.FlipH(img2)
	}

	rect := image.Rect(0, 0, img1.Bounds().Dx()*nlong+img2.Bounds().Dx(), img1.Bounds().Dy()*nrows)
	canvas := image.NewRGBA(rect)

	for row := 0; row < nrows; row++ {
		y := int(float64(img1.Bounds().Dy()*row) * rinterval)
		for i := 0; i < nlong; i++ {
			rect = image.Rect(img1.Bounds().Dx()*i, y, img1.Bounds().Dx()*(i+1), y+img1.Bounds().Dy())
			draw.Draw(canvas, rect, img1, image.ZP, draw.Over)
		}
		x := int(float64(img1.Bounds().Dx()*nlong))
		rect = image.Rect(x, y, x+img2.Bounds().Dx(), y+img1.Bounds().Dy())
		draw.Draw(canvas, rect, img2, image.ZP, draw.Over)
	}

	var output image.Image = canvas
	if flipV {
		output = imaging.FlipV(output)
	}

	if filename != "" {
		err = saveImage(filename, output)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	var buf bytes.Buffer
	err = sixel.NewEncoder(&buf).Encode(output)
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(buf.Bytes())
	os.Stdout.Sync()
}
