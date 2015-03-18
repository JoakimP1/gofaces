package main

import (
	"fmt"
	"log"
	"os"

	"flag"
	"github.com/gographics/imagick/imagick"
	//	nude "github.com/koyachi/go-nude"
	"bytes"
	"github.com/lazywei/go-opencv/opencv"
	gomat "github.com/skelterjohn/go.matrix"
	"image/jpeg"
	"math"
	"math/rand"
	"time"
)

const maxPixels = 100

var rowCount int
var colCount int

var filePath *string
var pictureFiles []string

var pixelMatrix [][]float64
var meanMatrix []float64
var diffMatrix [][]float64
var covMatrix [][]float64

var eigenMatrix [][]float64

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func GetFileNames(path string) []string {
	dir, err := os.Open(path)
	checkErr(err)
	defer dir.Close()
	fi, err := dir.Stat()
	checkErr(err)
	filenames := make([]string, 0)
	if fi.IsDir() {
		fis, err := dir.Readdir(-1) // -1 means return all the FileInfos
		checkErr(err)
		for _, fileinfo := range fis {
			if !fileinfo.IsDir() {
				filenames = append(filenames, path+fileinfo.Name())
			}
		}
	}
	return filenames
}

func NewMatrix(rows, cols int) [][]float64 {
	matrix := make([][]float64, rows, rows)
	for i := 0; i < rows; i++ {
		matrix[i] = make([]float64, cols, cols)
	}
	return matrix
}

func ComputeMeanColumn() {

	for k := 0; k < rowCount; k++ {
		sum := 0.0
		for l := 0; l < colCount; l++ {
			sum += pixelMatrix[k][l]
		}
		meanMatrix[k] = sum / float64(colCount)
	}

}

func ComputeDifferenceMatrixPixels() {
	for i := 0; i < rowCount; i++ {
		for j := 0; j < colCount; j++ {
			diffMatrix[i][j] = pixelMatrix[i][j] - meanMatrix[i]
		}
	}
}

func ComputeCovarianceMatrix() {
	for i := 0; i < colCount; i++ {
		for j := 0; j < colCount; j++ {
			sum := 0.0
			for k := 0; k < rowCount; k++ {
				sum += diffMatrix[k][i] * diffMatrix[k][j]
			}
			covMatrix[i][j] = sum
		}
	}
}

func createPicture() {
	newWand := imagick.NewPixelWand()
	defer newWand.Destroy()
	newWand.SetColor("white")
	newImg := imagick.NewMagickWand()
	defer newImg.Destroy()
	// Create a 100x100 image with a default of white
	newImg.NewImage(uint(maxPixels), uint(maxPixels), newWand)
	// Get a new pixel iterator
	iterator := newImg.NewPixelIterator()
	defer iterator.Destroy()
	var pixelIndex = 0
	for y := 0; y < int(newImg.GetImageHeight()); y++ {
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
			//gray := 255 * meanMatrix[pixelIndex]
			gray := 255 * meanMatrix[pixelIndex]
			fmt.Println("mean pixel:", pixelIndex, meanMatrix[pixelIndex], 255*meanMatrix[pixelIndex])

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
	newImg.WriteImage(*filePath + "jpg/mean.jpg")
}

func getPixelVectorFromFile(path string) []float64 {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()
	pixelVector := make([]float64, rowCount, rowCount)

	fmt.Println("filename:", path)
	err := mw.ReadImage(path)
	if err != nil {
		panic(err)
	}

	//Crop
	mw.AdaptiveResizeImage(uint(maxPixels), uint(maxPixels))

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

type pixelCoord struct {
	x      int
	y      int
	width  int
	height int
}

func (coord1 *pixelCoord) Distance(coord2 pixelCoord) int {
	dx := float64(coord2.x - coord1.x)
	dy := float64(coord2.y - coord1.y)
	return int(math.Sqrt(dx*dx + dy*dy))
}

func (c *pixelCoord) Center() pixelCoord {
	return pixelCoord{
		x: c.x + c.width/2,
		y: c.y + c.height/2,
	}
}

func (f *face) Center() pixelCoord {
	le := f.eye_left.Center()
	return pixelCoord{
		x: le.x + le.Distance(f.eye_right.Center())/2,
		y: le.y,
	}
}

type face struct {
	coord     pixelCoord
	eye_left  pixelCoord
	eye_right pixelCoord
}

func (face *face) Angle() float64 {

	r := math.Atan2(float64(face.eye_right.y-face.eye_left.y), float64(face.eye_right.x-face.eye_left.x))
	if r > 0.0 {
		return r
	} else {
		return (-r * 180) / math.Pi
	}

}

type faces []face

var debug = true

var eyeCascade *opencv.HaarCascade
var faceCascade *opencv.HaarCascade

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	log.SetFlags(log.Flags() | log.Llongfile)
	//currentDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	filePath = flag.String("f", "/home/joakim/Go/src/github.com/joakimp1/gofaces/", "Path to images.")
	flag.Parse()

	pictureFiles = GetFileNames(*filePath + "jpg/train2/")

	rowCount = maxPixels * maxPixels
	colCount = len(pictureFiles)

	pixelMatrix = NewMatrix(rowCount, colCount)
	meanMatrix = make([]float64, rowCount, rowCount)
	diffMatrix = NewMatrix(rowCount, colCount)
	covMatrix = NewMatrix(colCount, colCount)

	eigenMatrix = NewMatrix(rowCount, colCount)

	imagick.Initialize()
	defer imagick.Terminate()

	eyeCascade = opencv.LoadHaarClassifierCascade("/home/joakim/opencv-2.4.9/data/haarcascades/haarcascade_eye.xml")

	faceCascade = opencv.LoadHaarClassifierCascade("/home/joakim/Go/src/github.com/lazywei/go-opencv/samples/haarcascade_frontalface_alt.xml")

}

func main() {

	//image := opencv.LoadImage("/home/joakim/Go/src/github.com/joakimp1/gofaces/jpg/dump/3177685139_bcc3070261_z.jpg")
	image := opencv.LoadImage("/home/joakim/Go/src/github.com/joakimp1/gofaces/jpg/train2/BioID_0018.pgm")
	//image := opencv.LoadImage("/home/joakim/Go/src/github.com/joakimp1/gofaces/jpg/train2/BioID_0005.pgm")
	//image := opencv.LoadImage("/home/joakim/Go/src/github.com/joakimp1/gofaces/jpg/tantan/newfacesmall2.jpg")
	fmt.Println("image", image.Width(), image.Height())

	var faces = make(faces, 0)

	detFaces := faceCascade.DetectObjects(image)
	if len(detFaces) != 1 {
		if len(detFaces) > 1 {
			fmt.Println("To many faces in image")
		} else {
			fmt.Println("No faces found in image")
		}
		fmt.Println("exit")
		return
	}

	//var faceCrop *opencv.IplImage
	//faceCrop = opencv.Crop(image, int(float64(detFace.X())-(0.2*float64(detFace.X()))), int(float64(detFace.Y())-(0.2*float64(detFace.Y()))), int(float64(detFace.Width())+(0.2*float64(detFace.Width()))), int(float64(detFace.Height())+(0.2*float64(detFace.Height()))))

	for _, detFace := range detFaces {

		faceCoords := pixelCoord{
			x:      detFace.X(),
			y:      detFace.Y(),
			width:  detFace.Width(),
			height: detFace.Height(),
		}

		eyes := eyeCascade.DetectObjects(image)
		if len(eyes) < 2 {
			fmt.Println("Less than 2 eyes found")
			return
		}

		if debug {

			for key, eye := range eyes {

				fmt.Println("eye", key)

				opencv.Line(image,
					eye.Center(),
					eye.Center(),
					opencv.ScalarAll(255.0), 3, 1, 0)

				opencv.Rectangle(image,
					opencv.Point{eye.X() + eye.Width(), eye.Y()},
					opencv.Point{eye.X(), eye.Y() + eye.Height()},
					opencv.ScalarAll(255.0), 2, 1, 0)

			}
			fmt.Println("eyes info", eyes)
		}

		eye_1 := pixelCoord{
			x:      eyes[1].X(),
			y:      eyes[1].Y(),
			width:  eyes[1].Width(),
			height: eyes[1].Height(),
		}

		eye_2 := pixelCoord{
			x:      eyes[0].X(),
			y:      eyes[0].Y(),
			width:  eyes[0].Width(),
			height: eyes[0].Height(),
		}

		// sometimes eyes are inversed ! we switch them
		if eye_1.x < eye_2.x {
			faces = append(faces, face{
				coord:     faceCoords,
				eye_left:  eye_1,
				eye_right: eye_2,
			})
		} else {
			faces = append(faces, face{
				coord:     faceCoords,
				eye_left:  eye_2,
				eye_right: eye_1,
			})
		}

		//		fmt.Println("faces ", facekey, faces)

	}

	//faceCrop = opencv.Crop(faceCrop, int(float64(face.X())-(0.2*float64(face.X()))), int(float64(face.Y())-(0.2*float64(face.Y()))), int(float64(face.Width())+(0.2*float64(face.Width()))), int(float64(face.Height())+(0.2*float64(face.Height()))))

	fooo := image.ToImage()

	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, fooo, nil)

	mw := imagick.NewMagickWand()

	// Schedule cleanup
	defer mw.Destroy()
	err = mw.ReadImageBlob(buf.Bytes())
	if err != nil {
		panic(err)
	}

	center := faces[0].eye_left.Center()

	center2 := faces[0].Center()

	fmt.Println("Left Eye center:", center)
	fmt.Println("Face Eye center:", center2)

	srt := make([]float64, 7)
	srt[0] = float64(center.x)
	srt[1] = float64(center.y)
	srt[2] = float64(1)
	srt[3] = float64(1)
	srt[4] = faces[0].Angle()
	srt[5] = float64((image.Width() / 2) - (center2.x - center.x))
	srt[6] = float64(image.Height() / 2)

	//	srt := make([]float64, 7)
	//	srt[0] = float64(center2.x)
	//	srt[1] = float64(center.y)
	//	srt[2] = float64(1)
	//	srt[3] = float64(1)
	//	srt[4] = faces[0].Angle()
	//	srt[5] = float64(image.Width() / 2)
	//	srt[6] = float64(image.Height() / 2)

	//X,Y     Scale     Angle
	//X,Y     Scale     Angle  NewX,NewY
	//	srt[0] = float64(faces[0].eye_left.x)
	//	srt[1] = float64(faces[0].eye_left.y)
	//	srt[2] = 1.0
	//	srt[3] = faces[0].Angle() * 2

	mw.DistortImage(imagick.DISTORTION_SCALE_ROTATE_TRANSLATE, srt, false)
	fmt.Println("angle:", faces[0].Angle())
	fmt.Println("STR:", srt)

	mw.GetImagesBlob()

	win := opencv.NewWindow("Face Detection")
	defer win.Destroy()
	printImg := opencv.DecodeImageMem(mw.GetImagesBlob())
	if debug {

		opencv.Line(printImg,
			opencv.Point{0, printImg.Height() / 2},
			opencv.Point{printImg.Width(), printImg.Height() / 2},
			opencv.ScalarAll(255.0), 2, 1, 0)

		opencv.Line(printImg,
			opencv.Point{printImg.Width() / 2, 0},
			opencv.Point{printImg.Width() / 2, printImg.Height()},
			opencv.ScalarAll(55.0), 1, 1, 0)

	}

	win.ShowImage(printImg)
	opencv.WaitKey(0)

	return

	for i := 0; i < colCount; i++ {
		tempMatrix := getPixelVectorFromFile(pictureFiles[i])
		for j := 0; j < rowCount; j++ {
			pixelMatrix[j][i] = tempMatrix[j]
		}
	}
	ComputeMeanColumn()
	ComputeDifferenceMatrixPixels()
	ComputeCovarianceMatrix()
	ComputeEigenFaces()

	//	falsePictureFiles := GetFileNames(*filePath + "jpg/tantan/")
	//	for i := 0; i < len(falsePictureFiles); i++ {
	//		subjectMatrix := getPixelVectorFromFile(falsePictureFiles[i])
	//
	//		dist2 := ComputeDistance(subjectMatrix)
	//		fmt.Println("dist:", falsePictureFiles[i], dist2)
	//
	//	}
	//	falsePictureFiles2 := GetFileNames(*filePath + "jpg/train1/")
	//	for i := 0; i < len(falsePictureFiles2); i++ {
	//		subjectMatrix := getPixelVectorFromFile(falsePictureFiles2[i])
	//
	//		dist := ComputeDistance(subjectMatrix)
	//		fmt.Println("dist:", falsePictureFiles2[i], dist)
	//	}

	subjectMatrix := getPixelVectorFromFile(pictureFiles[4])
	dist := ComputeDistance(subjectMatrix)
	fmt.Println("dist:", pictureFiles[4], dist)

	subjectMatrix = getPixelVectorFromFile("/home/joakim/Go/src/github.com/joakimp1/gofaces/jpg/tantan/newfacesmall2.jpg")
	dist = ComputeDistance(subjectMatrix)
	fmt.Println("dist:", "/home/joakim/Go/src/github.com/joakimp1/gofaces/jpg/tantan/newfacesmall2.jpg", dist)

	subjectMatrix = getPixelVectorFromFile("/home/joakim/eigen/bioid/BioID_1213.pgm")
	dist = ComputeDistance(subjectMatrix)
	fmt.Println("dist:", "/home/joakim/eigen/bioid/BioID_1213.pgm", dist)

}

func ComputeDistance(subjectPixels []float64) float64 {
	diffPixels := ComputeDifferencePixels(subjectPixels)
	weights := ComputeWeights(diffPixels)
	reconstructedEigenPixels := ReconstructImageWithEigenFaces(weights)
	return ComputeImageDistance(subjectPixels, reconstructedEigenPixels)

}

func ComputeImageDistance(pixels1, pixels2 []float64) float64 {

	distance := 0.0
	for i := 0; i < rowCount; i++ {
		diff := pixels1[i] - pixels2[i]
		distance += diff * diff
	}

	return math.Sqrt(distance / float64(rowCount))
}

func ComputeDifferencePixels(subjectPixels []float64) (diffPixels []float64) {
	diffPixels = make([]float64, rowCount, rowCount)
	for i := 0; i < rowCount; i++ {
		diffPixels[i] = subjectPixels[i] - meanMatrix[i]
	}
	return
}

func ComputeWeights(diffImagePixels []float64) []float64 {
	eigenWeights := make([]float64, rowCount, rowCount)
	for i := 0; i < colCount; i++ {
		for j := 0; j < rowCount; j++ {
			eigenWeights[i] += diffImagePixels[j] * eigenMatrix[j][i]
		}
	}

	return eigenWeights
}

func ReconstructImageWithEigenFaces(weights []float64) []float64 {
	reconstructedPixels := make([]float64, rowCount, rowCount)

	for i := 0; i < colCount; i++ {
		for j := 0; j < rowCount; j++ {
			reconstructedPixels[j] += weights[i] * eigenMatrix[j][i]
		}
	}

	for i := 0; i < rowCount; i++ {
		reconstructedPixels[i] += meanMatrix[i]
	}

	min := float64(math.MaxFloat64)
	max := float64(-math.MaxFloat64)
	fmt.Println(min, max)

	for i := 0; i < rowCount; i++ {
		min = math.Min(min, reconstructedPixels[i])
		max = math.Max(max, reconstructedPixels[i])
	}
	fmt.Println(min, max)

	normalizedReconstructedPixels := make([]float64, rowCount, rowCount)
	for i := 0; i < rowCount; i++ {
		normalizedReconstructedPixels[i] = (255.0 * (reconstructedPixels[i] - min)) / (max - min)
	}

	return normalizedReconstructedPixels
}

func ComputeEigenFaces() {

	pixelCount := len(diffMatrix)

	denseMat := gomat.MakeDenseMatrixStacked(covMatrix)
	eigenVectors, _, _ := denseMat.Eigen()

	imageCount := eigenVectors.Cols()
	rank := eigenVectors.Rows()

	for i := 0; i < rank; i++ {
		sumSquare := 0.0
		for j := 0; j < pixelCount; j++ {
			for k := 0; k < imageCount; k++ {

				eigenMatrix[j][i] += diffMatrix[j][k] * eigenVectors.Get(i, k)
			}
			sumSquare += eigenMatrix[j][i] * eigenMatrix[j][i]
		}
		norm := math.Sqrt(float64(sumSquare))
		for j := 0; j < pixelCount; j++ {
			eigenMatrix[j][i] /= norm
		}
	}

}
