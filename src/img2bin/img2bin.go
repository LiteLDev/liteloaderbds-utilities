package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
)

type Pixel struct {
	R int
	G int
	B int
	A int
}

func getPixels(file io.Reader) (pixel [][]Pixel, e error, width, height int) {
	img, _, err := image.Decode(file)

	if err != nil {
		return nil, err, 0, 0
	}

	bounds := img.Bounds()
	width, height = bounds.Max.X, bounds.Max.Y
	if width%128 != 0 || height%128 != 0 {
		fmt.Println("[Warn]\tImage Size Error, should be Integral multiple of 128 * 128 ")
	}
	var pixels [][]Pixel
	for y := 0; y < height; y++ {
		var row []Pixel
		for x := 0; x < width; x++ {
			row = append(row, rgbaToPixel(img.At(x, y).RGBA()))
		}
		pixels = append(pixels, row)
	}

	return pixels, nil, width, height
}

func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	return Pixel{int(r / 257), int(g / 257), int(b / 257), int(a / 257)}
}

func main() {
	fmt.Println("[Info]\tImg2Bin For Liteloader&BedrockX Map Plugin, By WangYneos")
	fmt.Println("[Info]\tSupported Img Format : .png .jpeg")
	var input, output string
	flag.StringVar(&input, "in", "image.png", "Input Image Path")
	flag.StringVar(&output, "out", "map", "Out Binary Path")
	flag.Parse()

	//fmt.Printf("%v\n",*ip)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	imageFile, err := os.Open(input)
	if err != nil {
		fmt.Println("[Error]\tNo Such File >", input)
		os.Exit(-1)
	}
	fmt.Println("[Info]\tInput Image Loaded >", input)
	defer imageFile.Close()

	pixels, err, width, height := getPixels(imageFile)
	if err != nil {
		fmt.Println("[Error] Can't Decode The Image,make sure you are using png or jpeg")
		os.Exit(-1)
	}
	fmt.Printf("%s%s%s\n", "[Info]\tImage Decoded, Output to Binary: ", output, "-w_h")

	//write.WriteString(str)
	//write.Write()
	fmt.Println("[Info]\tImageSize > ", height, "*", width)
	// Spilt into  chunks
	wcut := width / 128
	hcut := height / 128
	if width%128 != 0 {
		wcut++
	}

	if height%128 != 0 {
		hcut++
	}
	fmt.Printf("[Info]\tCut To Muilt Files > width=%d height=%d\n", wcut, hcut)
	for h := 0; h < hcut; h++ {
		for w := 0; w < wcut; w++ {
			fmt.Printf("(%d,%d) [%d,%d]\t", w, h, w*128, h*128)
			OutputBin, err := os.OpenFile(fmt.Sprintf("%s-%d_%d", output, w, h), os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				fmt.Println("[Error]\tCan't Create&Open File", err)
				os.Exit(-1)
			}

			defer OutputBin.Close()
			write := bufio.NewWriter(OutputBin)
			for a := h * 128; a < h*128+128; a++ {
				for b := w * 128; b < w*128+128; b++ {
					if a < len(pixels) && b < len(pixels[a]) {
						//WRITE RGBA
						write.Write([]byte{byte(pixels[a][b].R), byte(pixels[a][b].G), byte(pixels[a][b].B), byte(pixels[a][b].A)})
					} else {
						//fill empty chunk
						write.Write([]byte{byte(0xff), byte(0xff), byte(0xff), byte(00)})
					}
				}
			}
			//chunk finished
			write.Flush()

		}
		fmt.Println("")
	}

	fmt.Println("[Info]\tDone！")
}
