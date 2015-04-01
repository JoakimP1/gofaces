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

func GetNormalizedCroppedByteVectorFromFile(width, height int, path string) []byte {

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

//func GetImageFromFile(path string) []byte {
//	file, err := os.Open(path)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer file.Close()
//	img, err := pgm.Decode(file)
//	if err != nil {
//		log.Fatal(os.Stderr, "%s: %v\n", "./selfcss.png", err)
//	}
//}

func GetByteVectorFromFile(path string) []byte {

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	err := mw.ReadImage(path)
	if err != nil {
		panic(err)
	}
	mw.SetImageFormat("PGM")
	return mw.GetImageBlob()
}

func Flop(img []byte) []byte {

	mw := imagick.NewMagickWand()
	defer mw.Destroy()


	err := mw.ReadImageBlob(img)
	if err != nil {
		panic(err)
	}
	mw.FlopImage()
	return mw.GetImageBlob()
}

func GetNormalizedByteVectorFromFile(path string) []byte {

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	err := mw.ReadImage(path)
	mw.FlopImage()

	if err != nil {
		panic(err)
	}
	mw.SetImageFormat("JPEG")

	mw.TransformImageColorspace(imagick.COLORSPACE_RGB)
	mw.SeparateImageChannel(1)
	mw.LevelImage(0, 2, 65535)

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

func CropOutFace(img []byte, face face) []byte {
	mw := imagick.NewMagickWand()
	// Schedule cleanup
	defer mw.Destroy()

	err := mw.ReadImageBlob(img)
	if err != nil {
		panic(err)
	}

	newWidth := face.coord.Width()
	newHeight := face.coord.Height()

	newX := (int(mw.GetImageWidth()) / 2) - (newWidth / 2)
	newY := (int(mw.GetImageHeight()) / 2) - (newHeight / 2)

	mw.CropImage(uint(newWidth), uint(float64(newHeight)*1.2), newX, newY)
	return mw.GetImageBlob()
}

func AlignFaceInImage(img []byte, face face) []byte {

	mw := imagick.NewMagickWand()
	// Schedule cleanup
	defer mw.Destroy()
	err := mw.ReadImageBlob(img)
	if err != nil {
		panic(err)
	}

	center := face.eye_left.Center()
	newCenter := face.eye_left.Center().x - face.Center().x
	//fmt.Println("Face Center: ", center2, "Left eye Center: ", center)

	//X,Y ScaleX,ScaleY Angle  NewX,NewY
	srt := make([]float64, 7)

	//X 149.5
	srt[0] = float64(center.x) + 0.5
	//Y 160.5
	srt[1] = float64(center.y) + 0.5
	//Scale
	srt[2] = float64(1)
	srt[3] = float64(1)
	//Angle 8.21920924889906
	srt[4] = face.Angle()
	//NewX 147
	srt[5] = float64((int(mw.GetImageWidth())/2)+newCenter) + 0.5
	//NewY 192
	srt[6] = float64(int(mw.GetImageHeight()) / 2)

	//fmt.Println(srt)
	mw.DistortImage(imagick.DISTORTION_SCALE_ROTATE_TRANSLATE, srt, false)

	return mw.GetImageBlob()
}
