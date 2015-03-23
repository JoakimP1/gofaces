package gofaces

import (
	"fmt"
	"github.com/gographics/imagick/imagick"
)

type PixelVector []float64
type PixelBuffer []byte

//func ( pixels *PixelVector) {
//
//}

func PixelVectorToImage(imgVector []float64, width, height int) []byte {

	pw := imagick.NewPixelWand()
	defer pw.Destroy()
	pw.SetColor("white")

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Create a 100x100 image with a default of white
	mw.NewImage(uint(width), uint(height), pw)
	mw.SetImageFormat("jpeg")

	// Get a new pixel iterator
	iterator := mw.NewPixelIterator()
	defer iterator.Destroy()
	pixelIndex := 0
	for y := 0; y < int(mw.GetImageHeight()); y++ {
		// Get the next row of the image as an array of PixelWands
		pixels := iterator.GetNextIteratorRow()
		if len(pixels) == 0 {
			break
		}
		// Set the row of wands to a simple gray scale gradient
		for _, pixel := range pixels {
			if !pixel.IsVerified() {
				panic("unverified pixel")
			}
			gray := imgVector[pixelIndex]

			hex := fmt.Sprintf("#%02x%02x%02x", int(gray), int(gray), int(gray))
			if ret := pixel.SetColor(hex); !ret {
				panic("Could not set color in pixel")
			}
			pixelIndex++
		}

		// Sync writes the pixels back to the mw
		if err := iterator.SyncIterator(); err != nil {
			panic(err)
		}
	}
	return mw.GetImageBlob()
}

func CreatePictureFromVector(imgVector []float64, width, height int, path string) {
	newWand := imagick.NewPixelWand()
	defer newWand.Destroy()
	newWand.SetColor("white")
	newImg := imagick.NewMagickWand()
	defer newImg.Destroy()

	// Create a nXn image with a default of white
	newImg.NewImage(uint(width), uint(height), newWand)
	// Get a new pixel iterator
	iterator := newImg.NewPixelIterator()
	defer iterator.Destroy()
	var pixelIndex = 0
	for y := 0; y < height; y++ {
		// Get the next row of the image as an array of PixelWands
		pixels := iterator.GetNextIteratorRow()
		if len(pixels) == 0 {
			break
		}
		// Set the row of wands to a simple gray scale gradient
		for _, pixel := range pixels {
			if !pixel.IsVerified() {
				panic("unverified pixel")
			}
			gray := imgVector[pixelIndex]

			hex := fmt.Sprintf("#%02x%02x%02x", int(gray), int(gray), int(gray))
			if ret := pixel.SetColor(hex); !ret {
				panic("Could not set color in pixel")
			}
			pixelIndex++
		}
		// Sync writes the pixels back to the mw
		if err := iterator.SyncIterator(); err != nil {
			panic(err)
		}
	}
	newImg.WriteImage(path)
}

func GetNormalizedByteVectorFromFile(width, height int, path string) []byte {

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	err := mw.ReadImage(path)
	if err != nil {
		panic(err)
	}

	//Crop
	mw.AdaptiveResizeImage(uint(width), uint(height))

	//Normalize it
	mw.TransformImageColorspace(imagick.COLORSPACE_RGB)
	mw.SeparateImageChannel(1)
	mw.LevelImage(0, 2, 65535)
	//	mw.NormalizeImage()
	//	mw.AutoLevelImage()

	return mw.GetImageBlob()
}

func GetByteVectorFromFile(path string) []byte {

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	err := mw.ReadImage(path)
	if err != nil {
		panic(err)
	}
	mw.SetImageFormat("jpeg")
	return mw.GetImageBlob()
}

func GetNormalizedPixelVectorFromBuffer(width, height int, img []byte) []float64 {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()
	pixelVector := make([]float64, width*height, width*height)

	err := mw.ReadImageBlob(img)
	if err != nil {
		panic(err)
	}

	//Crop
	mw.AdaptiveResizeImage(uint(width), uint(height))

	//Normalize it
	mw.TransformImageColorspace(imagick.COLORSPACE_RGB)
	mw.SeparateImageChannel(1)
	mw.LevelImage(0, 2, 65535)
	//mw.NormalizeImage()
	//mw.AutoLevelImage()

	// Get a new pixel iterator
	iterator := mw.NewPixelIterator()
	defer iterator.Destroy()
	var pixelIndex = 0
	for y := 0; y < int(mw.GetImageHeight()); y++ {

		// Get the next row of the image as an array of PixelWands
		pixels := iterator.GetNextIteratorRow()
		if len(pixels) == 0 {
			break
		}

		for _, pixel := range pixels {
			if !pixel.IsVerified() {
				panic("unverified pixel")
			}

			pixelVector[pixelIndex] = 255.0 * pixel.GetRed()
			pixelIndex++
		}
	}

	return pixelVector
}

func GetNormalizedPixelVectorFromFile(width, height int, path string) []float64 {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()
	pixelVector := make([]float64, width*height, width*height)

	err := mw.ReadImage(path)
	if err != nil {
		panic(err)
	}

	//Crop
	mw.AdaptiveResizeImage(uint(width), uint(height))

	//Normalize it
	mw.TransformImageColorspace(imagick.COLORSPACE_RGB)
	mw.SeparateImageChannel(1)
	mw.LevelImage(0, 2, 65535)
	//mw.NormalizeImage()
	//mw.AutoLevelImage()

	// Get a new pixel iterator
	iterator := mw.NewPixelIterator()
	defer iterator.Destroy()
	var pixelIndex = 0
	for y := 0; y < int(mw.GetImageHeight()); y++ {

		// Get the next row of the image as an array of PixelWands
		pixels := iterator.GetNextIteratorRow()
		if len(pixels) == 0 {
			break
		}

		for _, pixel := range pixels {
			if !pixel.IsVerified() {
				panic("unverified pixel")
			}

			pixelVector[pixelIndex] = 255.0 * pixel.GetRed()
			pixelIndex++
		}
	}

	return pixelVector
}

func alignFaceInImage(img *[]byte, face face) []byte {

	mw := imagick.NewMagickWand()
	// Schedule cleanup
	defer mw.Destroy()
	err := mw.ReadImageBlob(*img)
	if err != nil {
		panic(err)
	}

	center := face.eye_left.Center()
	center2 := face.Center()

	srt := make([]float64, 7)
	srt[0] = float64(center.x)
	srt[1] = float64(center.y)
	srt[2] = float64(1)
	srt[3] = float64(1)
	srt[4] = face.Angle()
	srt[5] = float64((int(mw.GetImageWidth()) / 2) - int(center2.x-center.x))
	srt[6] = float64(int(mw.GetImageWidth()) / 2)

	mw.DistortImage(imagick.DISTORTION_SCALE_ROTATE_TRANSLATE, srt, false)

	return mw.GetImageBlob()

}
