package gofaces

import (
	"bytes"
	"fmt"
	"github.com/lazywei/go-opencv/opencv"

	"image/jpeg"
	"math"
)

type pixelCoord struct {
	x      int
	y      int
	width  int
	height int
}

type pixelArea interface {
	X() int
	Y() int
	Width() int
	Height() int
}

func (coord1 *pixelCoord) Distance(coord2 pixelCoord) int {
	dx := float64(coord2.x - coord1.x)
	dy := float64(coord2.y - coord1.y)
	return int(math.Sqrt(dx*dx + dy*dy))
}

func (coord *pixelCoord) X() int {
	return coord.x
}

func (coord *pixelCoord) Y() int {
	return coord.y
}

func (coord *pixelCoord) Width() int {
	return coord.width
}

func (coord *pixelCoord) Height() int {
	return coord.height
}

func (c *pixelCoord) Center() pixelCoord {
	return pixelCoord{
		x: c.x + c.width/2,
		y: c.y + c.height/2,
	}
}

func (c *pixelCoord) Flop(coord2 pixelCoord) {
	c.x = coord2.Width() - c.X() - c.Width()
}

func (f *face) Flop(width, height int) {
	flopAround := pixelCoord{width: width, height: height}
	f.coord.Flop(flopAround)
	f.eye_left.Flop(flopAround)
	f.eye_right.Flop(flopAround)
}

func (f *face) Width() int {
	return f.coord.width
}

func (f *face) Center() pixelCoord {
	le := f.eye_left.Center()
	return pixelCoord{
		x: le.x + le.Distance(f.eye_right.Center())/2,
		y: le.y,
	}
}

func (f *face) Eyes() int {
	eyes := 0
	if f.eye_left.Width() > 0 && f.eye_left.Height() > 0 {
		eyes++
	}
	if f.eye_right.Width() > 0 && f.eye_right.Height() > 0 {
		eyes++
	}
	return eyes
}

func (f *face) DistanceBetweenEyes() int {
	le := f.eye_left.Center()
	re := f.eye_right.Center()
	return re.x - le.x
}

func (f *face) LeftEye() *pixelCoord {
	return &f.eye_left
}

func (f *face) RightEye() *pixelCoord {
	return &f.eye_right
}

type face struct {
	coord     pixelCoord
	eye_left  pixelCoord
	eye_right pixelCoord
	mouth     pixelCoord
	nose      pixelCoord
}

type faces []face

func (face *face) Angle() float64 {

	r := math.Atan2(float64(face.eye_right.Center().y-face.eye_left.Center().y), float64(face.eye_right.Center().x-face.eye_left.Center().x))

	return (-r * 180) / math.Pi
	//	if r > 0.0 {
	//		return r
	//	} else {
	//		//		return (-r * 360) / math.Pi
	//		return (-r * 180) / math.Pi
	//	}
}

type FaceDetector struct {
	eyeCascade      *opencv.HaarCascade
	lefteyeCascade  *opencv.HaarCascade
	righteyeCascade *opencv.HaarCascade
	faceCascade     *opencv.HaarCascade
	mouthCascade    *opencv.HaarCascade
	noseCascade     *opencv.HaarCascade
}

func NewFaceDetector() *FaceDetector {

	detector := &FaceDetector{
		eyeCascade:      opencv.LoadHaarClassifierCascade("/home/joakim/opencv-2.4.9/data/haarcascades/haarcascade_eye.xml"),
		lefteyeCascade:  opencv.LoadHaarClassifierCascade("/home/joakim/opencv-2.4.9/data/haarcascades/haarcascade_mcs_lefteye.xml"),
		righteyeCascade: opencv.LoadHaarClassifierCascade("/home/joakim/opencv-2.4.9/data/haarcascades/haarcascade_mcs_righteye.xml"),
		faceCascade:     opencv.LoadHaarClassifierCascade("/home/joakim/Go/src/github.com/lazywei/go-opencv/samples/haarcascade_frontalface_alt.xml"),
		mouthCascade:    opencv.LoadHaarClassifierCascade("/home/joakim/opencv-2.4.9/data/haarcascades/haarcascade_mcs_mouth.xml"),
		noseCascade:     opencv.LoadHaarClassifierCascade("/home/joakim/opencv-2.4.9/data/haarcascades/haarcascade_mcs_nose.xml"),
	}
	return detector
}

func (detector *FaceDetector) DetectEyes(image *opencv.IplImage, roi pixelArea) (leftEyes, rightEyes []pixelCoord) {

	var topFaceLeft, topFaceRight opencv.Rect

	topFaceLeft.Init(roi.X(), roi.Y()+int(float64(roi.Height())*0.20), roi.Width()/2, int(float64(roi.Height())/2))

	topFaceRight.Init(roi.X()+roi.Width()/2, roi.Y()+int(float64(roi.Height())*0.20), roi.Width()/2, int(float64(roi.Height())/2))

	fmt.Println("topFaceLeft", image.Width(), image.Height(), topFaceLeft)

	image.SetROI(topFaceLeft)
	for _, eye := range detector.eyeCascade.DetectObjects(image) {
		leftEyes = append(leftEyes, pixelCoord{
			x:      eye.X() + topFaceLeft.X(),
			y:      eye.Y() + topFaceLeft.Y(),
			width:  eye.Width(),
			height: eye.Height(),
		})
	}
	fmt.Println(len(leftEyes), " left eyes found")

	image.SetROI(topFaceRight)
	for _, eye := range detector.righteyeCascade.DetectObjects(image) {
		rightEyes = append(rightEyes, pixelCoord{
			x:      eye.X() + topFaceRight.X(),
			y:      eye.Y() + topFaceRight.Y(),
			width:  eye.Width(),
			height: eye.Height(),
		})
	}
	fmt.Println(len(rightEyes), " right eyes found")
	image.ResetROI()
	return
}

func FindBestEyes(leftEyes, rightEyes []pixelCoord) (leftEye, rightEye pixelCoord) {

	var bestEye, bestDist, dist int

	if len(rightEyes) == 1 && len(leftEyes) == 1 {
		return leftEyes[0], rightEyes[0]

	} else if len(rightEyes) > 1 && len(leftEyes) == 1 {
		//Left Eye Found, right Eye multiple eyes
		leftEye = leftEyes[0]
		//Find the best eye with the lowest distance

		for key, eye := range rightEyes {
			dist = leftEye.Distance(eye)
			fmt.Println("Dist", dist)

			if dist < bestDist || bestDist == 0 {
				bestEye = key
				bestDist = dist
			}
			fmt.Println("bestDist", bestDist)
			fmt.Println("bestEye", bestEye)

		}
		fmt.Println("bestEye", rightEyes[bestEye])

		rightEye = rightEyes[bestEye]

	} else if len(rightEyes) == 1 && len(leftEyes) > 1 {
		//Right Eye Found, left Eye multiple eyes
		rightEye = rightEyes[0]
		//Find the best eye with the lowest distance

		for key, eye := range leftEyes {
			dist = rightEye.Distance(eye)

			if dist < bestDist || bestDist == 0 {
				bestEye = key
				bestDist = dist
			}
		}
		leftEye = leftEyes[bestEye]

	}

	return leftEye, rightEye
}

func (detector *FaceDetector) DetectFaces(image *opencv.IplImage) []pixelCoord {

	var faceCoords = make([]pixelCoord, 0)

	detFaces := detector.faceCascade.DetectObjects(image)

	fmt.Println(len(detFaces), " faces found")

	for _, detFace := range detFaces {

		faceCoords = append(faceCoords, pixelCoord{
			x:      detFace.X(),
			y:      detFace.Y(),
			width:  detFace.Width(),
			height: detFace.Height(),
		})
	}
	return faceCoords
}

func (detector *FaceDetector) DetectFacialFeatures(image *opencv.IplImage, faceCoords []pixelCoord) faces {

	var faces = make(faces, 0)
 
	for _, faceCoord := range faceCoords {

		leftEyes, rightEyes := detector.DetectEyes(image, &faceCoord)

		if len(leftEyes) == 0 || len(rightEyes) == 0 {

			//No eyes found, lets flop the picture and check again
			fmt.Println("Flopping Picture to look for eyes")

			floppedImage := opencv.DecodeImageMem(FlopImage(ToByteBuffer(image)))

			floppedFaceCoords := faceCoord

			floppedFaceCoords.Flop(pixelCoord{width: image.Width(), height: image.Height()})

			leftEyes, rightEyes = detector.DetectEyes(floppedImage, &floppedFaceCoords)

			if len(leftEyes) == 0 || len(rightEyes) == 0 {
				faces = append(faces, face{
					coord:     faceCoord,
					eye_left:  pixelCoord{},
					eye_right: pixelCoord{},
				})
			}

			leftEye, rightEye := FindBestEyes(leftEyes, rightEyes)
			leftEye.Flop(pixelCoord{width: image.Width(), height: image.Height()})
			rightEye.Flop(pixelCoord{width: image.Width(), height: image.Height()})

			faces = append(faces, face{
				coord:     faceCoord,
				eye_left:  rightEye,
				eye_right: leftEye,
			})

		} else {

			leftEye, rightEye := FindBestEyes(leftEyes, rightEyes)
			faces = append(faces, face{
				coord:     faceCoord,
				eye_left:  leftEye,
				eye_right: rightEye,
			})
		}
	}
	fmt.Println(faces)
	return faces
}

func ToByteBuffer(image *opencv.IplImage) []byte {

	buf := new(bytes.Buffer)

	err := jpeg.Encode(buf, image.ToImage(), nil)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func (detector *FaceDetector) Detect(img []byte) faces {

	image := opencv.DecodeImageMem(img)
	faceAreas := detector.DetectFaces(image)
	faces := detector.DetectFacialFeatures(image, faceAreas)
	return faces
}

func Crosshair(img []byte) []byte {

	image := opencv.DecodeImageMem(img)

	//Horizontal line

	opencv.Line(image,
		opencv.Point{0, image.Height() / 2},
		opencv.Point{image.Width(), image.Height() / 2},
		opencv.ScalarAll(0), 1, 1, 0)
	//vertical line

	opencv.Line(image,
		opencv.Point{image.Width() / 2, 0},
		opencv.Point{image.Width() / 2, image.Height()},
		opencv.ScalarAll(0), 1, 1, 0)

	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, image.ToImage(), nil)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func PaintFace(img []byte, face face) []byte {

	image := opencv.DecodeImageMem(img)

	opencv.Rectangle(image,
		opencv.Point{face.coord.x + face.coord.width, face.coord.y},
		opencv.Point{face.coord.x, face.coord.y + face.coord.height},
		opencv.ScalarAll(0), 1, 1, 0)

	le := face.eye_left.Center()
	re := face.eye_right.Center()

	opencv.Circle(image,
		opencv.Point{le.x, le.y},
		2,
		opencv.ScalarAll(255), 1, 1, 0)

	opencv.Circle(image,
		opencv.Point{re.x, re.y},
		2,
		opencv.ScalarAll(255), 1, 1, 0)

	faceCenter := face.Center()
	opencv.Circle(image,
		opencv.Point{faceCenter.x, faceCenter.y},
		2,
		opencv.ScalarAll(50), 1, 1, 0)

	opencv.Rectangle(image,
		opencv.Point{face.mouth.x + face.mouth.width, face.mouth.y},
		opencv.Point{face.mouth.x, face.mouth.y + face.mouth.height},
		opencv.ScalarAll(2), 1, 1, 0)

	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, image.ToImage(), nil)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func PaintFaces(img []byte, faces faces) []byte {

	image := opencv.DecodeImageMem(img)

	for _, face := range faces {

		opencv.Rectangle(image,
			opencv.Point{face.coord.x + face.coord.width, face.coord.y},
			opencv.Point{face.coord.x, face.coord.y + face.coord.height},
			opencv.ScalarAll(0), 1, 1, 0)

		le := face.LeftEye().Center()
		re := face.RightEye().Center()

		opencv.Circle(image,
			opencv.Point{le.X(), le.Y()},
			2,
			opencv.ScalarAll(255), 1, 1, 0)

		opencv.Circle(image,
			opencv.Point{re.x, re.y},
			2,
			opencv.ScalarAll(255), 1, 1, 0)

		faceCenter := face.Center()
		opencv.Circle(image,
			opencv.Point{faceCenter.x, faceCenter.y},
			2,
			opencv.ScalarAll(50), 1, 1, 0)

		//		opencv.Rectangle(image,
		//			opencv.Point{face.mouth.x + face.mouth.width, face.mouth.y},
		//			opencv.Point{face.mouth.x, face.mouth.y + face.mouth.height},
		//			opencv.ScalarAll(2), 1, 1, 0)
	}

	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, image.ToImage(), nil)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}
