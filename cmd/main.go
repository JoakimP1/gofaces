package main

import (
	"flag"
	"github.com/gographics/imagick/imagick"
	"github.com/joakimp1/gofaces"
	"log"
	"os"
	//	nude "github.com/koyachi/go-nude"

	"fmt"
	"github.com/lazywei/go-opencv/opencv"
	"math/rand"
	"time"
)

var filePath *string
var pictureFiles []string

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

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	log.SetFlags(log.Flags() | log.Llongfile)
	//currentDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	filePath = flag.String("f", "/home/joakim/Go/src/github.com/joakimp1/gofaces/", "Path to images.")
	flag.Parse()

	imagick.Initialize()
	defer imagick.Terminate()

}

func main() {

	width := 100
	height := 100

	//	pictureFiles = GetFileNames(*filePath + "jpg/train1/")
	//
	images := make([][]byte, 10, width*height)
	//
	//	for i := 0; i < len(pictureFiles); i++ {
	//		images[i] = gofaces.GetByteVectorFromFile(pictureFiles[i])
	//	}

	images[0] = gofaces.GetNormalizedByteVectorFromFile("/home/joakim/Go/src/github.com/joakimp1/gofaces/jpg/tantan/newfacesmall2.jpg")
	faces := gofaces.Detect(images[0])

	foo := gofaces.PaintFace(images[0], faces[0])

	//foo = gofaces.AlignFaceInImage(foo, faces[0])

	fmt.Println(faces[0], faces[0].DistanceBetweenEyes(), faces[0].Width(), float64(faces[0].DistanceBetweenEyes())/float64(faces[0].Width()))

	win := opencv.NewWindow("Face Detection")
	defer win.Destroy()

	win.ShowImage(opencv.DecodeImageMem(foo))
	opencv.WaitKey(0)

	//
	//
	//
	//
	//
	//
	//
	//	for i := 0; i < colCount; i++ {
	//		tempMatrix := getPixelVectorFromFile(falsePictureFiles[i])
	//		for j := 0; j < rowCount; j++ {
	//			pixelMatrix[j][i] = tempMatrix[j]
	//		}
	//	}
	//
	//	buf := new(bytes.Buffer)
	//	err := jpeg.Encode(buf, fooo, nil)
	//
	//	eigenFace := NewEigenFace( len( pictureFiles ), maxPixels*maxPixels, )

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
	//
	//	subjectMatrix := getPixelVectorFromFile(pictureFiles[4])
	//	dist := ComputeDistance(subjectMatrix)
	//	fmt.Println("dist:", pictureFiles[4], dist)
	//
	//	subjectMatrix = getPixelVectorFromFile("/home/joakim/Go/src/github.com/joakimp1/gofaces/jpg/tantan/newfacesmall2.jpg")
	//	dist = ComputeDistance(subjectMatrix)
	//	fmt.Println("dist:", "/home/joakim/Go/src/github.com/joakimp1/gofaces/jpg/tantan/newfacesmall2.jpg", dist)
	//
	//	subjectMatrix = getPixelVectorFromFile("/home/joakim/eigen/bioid/BioID_1213.pgm")
	//	dist = ComputeDistance(subjectMatrix)
	//	fmt.Println("dist:", "/home/joakim/eigen/bioid/BioID_1213.pgm", dist)
	//
	//	subjectMatrix := getPixelVectorFromFile(pictureFiles[4])
	//	dist := ComputeDistance(subjectMatrix)
	//	fmt.Println("dist:", pictureFiles[4], dist)
	//
	//	subjectMatrix = getPixelVectorFromFile("/home/joakim/Go/src/github.com/joakimp1/gofaces/jpg/tantan/newfacesmall2.jpg")
	//	dist = ComputeDistance(subjectMatrix)
	//	fmt.Println("dist:", "/home/joakim/Go/src/github.com/joakimp1/gofaces/jpg/tantan/newfacesmall2.jpg", dist)
	//
	//	subjectMatrix = getPixelVectorFromFile("/home/joakim/eigen/bioid/BioID_1213.pgm")
	//	dist = ComputeDistance(subjectMatrix)
	//	fmt.Println("dist:", "/home/joakim/eigen/bioid/BioID_1213.pgm", dist)

}
