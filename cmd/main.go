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
	"path/filepath"
	"strings"
	"time"
)

var filePath *string
var pictureFiles []string

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func isPicture(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))

	if ext == ".jpg" {
		return true
	} else if ext == ".jpeg" {
		return true
	} else if ext == ".pgm" {
		return true
	}

	return false
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
			if !fileinfo.IsDir() && isPicture(fileinfo.Name()) {
				filenames = append(filenames, path+"/"+fileinfo.Name())
			}
		}
	}
	return filenames
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	log.SetFlags(log.Flags() | log.Llongfile)
	//currentDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	filePath = flag.String("f", "/home/joakim/Go/src/github.com/joakimp1/gofaces", "Path to images.")
	flag.Parse()

	imagick.Initialize()
	defer imagick.Terminate()

}

func loop(path string, paint, crosshair bool) {

	pictureFiles = GetFileNames(path)

	//	images := make([][]byte, len(pictureFiles))
	faceDetector := gofaces.NewFaceDetector()

	//	for i := 0; i < len(pictureFiles); i++ {

	//	}

	fmt.Println("Pictures: ", len(pictureFiles))

	for i := 0; i < len(pictureFiles); i++ {

		fmt.Println("Processing picture: ", pictureFiles[i])
		img := gofaces.GetNormalizedByteVectorFromFile(pictureFiles[i])
		faces := faceDetector.Detect(img)
		if len(faces) > 0 {
			if paint {
				img = gofaces.PaintFace(img, faces[0])
			}

			if faces[0].Eyes() > 1 {
				img = gofaces.AlignFaceInImage(img, faces[0])
			}

			if crosshair {
				img = gofaces.Crosshair(img)
			}

			cropface := gofaces.CropOutFace(img, faces[0])

			win := opencv.NewWindow("Face Detected")

			win.ShowImage(opencv.DecodeImageMem(cropface))
			opencv.WaitKey(0)
			win.Destroy()
		} else {
			win := opencv.NewWindow("No face found")

			win.ShowImage(opencv.DecodeImageMem(img))
			opencv.WaitKey(0)
			win.Destroy()
		}

	}
}

func one(path string, paint bool) {
	faceDetector := gofaces.NewFaceDetector()

	picture := gofaces.GetNormalizedByteVectorFromFile(path)
	faces := faceDetector.Detect(picture)
	if paint {
		picture = gofaces.PaintFace(picture, faces[0])
	}
	picture = gofaces.AlignFaceInImage(picture, faces[0])
	picture = gofaces.CropOutFace(picture, faces[0])

	win := opencv.NewWindow("Face Detection")
	defer win.Destroy()

	win.ShowImage(opencv.DecodeImageMem(picture))
	opencv.WaitKey(0)
}

func debug(path string) {
	faceDetector := gofaces.NewFaceDetector()

	picture := gofaces.GetNormalizedByteVectorFromFile(path)

	image := opencv.DecodeImageMem(picture)
	faceAreas := faceDetector.DetectFaces(image)
	faces := faceDetector.DetectFacialFeatures(image, faceAreas)

	picture = gofaces.PaintFaces(picture, faces)

	win := opencv.NewWindow("Debug Face Detection")
	defer win.Destroy()

	win.ShowImage(opencv.DecodeImageMem(picture))
	opencv.WaitKey(0)
}

func main() {

	//	/home/joakim/Go/src/github.com/joakimp1/gofaces/jpg/train1/5.jpg
	//
	//one("/home/joakim/Go/src/github.com/joakimp1/gofaces/jpg/train1/16.jpg", true)
	//one("/home/joakim/Go/src/github.com/joakimp1/gofaces/jpg/train1/18.jpg", true)
	//one("/home/joakim/eigen/bioid/BioID_0937.pgm", true)

	pictureFiles = GetFileNames("/home/joakim/eigen/bioid/")
	for i := 0; i < len(pictureFiles); i++ {
		debug(pictureFiles[i])
	}

	//one("/home/joakim/Go/src/github.com/joakimp1/gofaces/jpg/train1/16.jpg", true)
	//loop(*filePath+"jpg/train1/", true, true)
	//loop("/home/joakim/eigen/bioid", true, true)
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
