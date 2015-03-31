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

func (f *face) DistanceBetweenEyes() int {
	le := f.eye_left.Center()
	re := f.eye_right.Center()
	return re.x - le.x
}

func (f *face) LeftEye() pixelCoord {
	return f.eye_left
}

func (f *face) RightEye() pixelCoord {
	return f.eye_right
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
	if r > 0.0 {
		return r
	} else {
		//		return (-r * 360) / math.Pi
		return (-r * 180) / math.Pi
	}
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

func (detector *FaceDetector) DetectFaces(image *opencv.IplImage) faces {

	var faces = make(faces, 0)

	detFaces := detector.faceCascade.DetectObjects(image)

	if len(detFaces) != 1 {
		if len(detFaces) > 1 {
			//			panic("To many faces in image")
		} else {
			//			panic("No faces found in image")
		}
	}
	fmt.Println(len(detFaces), " faces found")

	for _, detFace := range detFaces {

		faceCoords := pixelCoord{
			x:      detFace.X(),
			y:      detFace.Y(),
			width:  detFace.Width(),
			height: detFace.Height(),
		}

		var topFaceLeft opencv.Rect
		topFaceLeft.Init(detFace.X(), detFace.Y()+int(float64(detFace.Height())*0.20), detFace.Width()/2, int(float64(detFace.Height())/2))

		var topFaceRight opencv.Rect
		topFaceRight.Init(detFace.X()+detFace.Width()/2, detFace.Y()+int(float64(detFace.Height())*0.20), detFace.Width()/2, int(float64(detFace.Height())/2))

		image.SetROI(topFaceLeft)
		lefteye := detector.eyeCascade.DetectObjects(image)

		eye_1 := pixelCoord{
			x:      lefteye[0].X() + topFaceLeft.X(),
			y:      lefteye[0].Y() + topFaceLeft.Y(),
			width:  lefteye[0].Width(),
			height: lefteye[0].Height(),
		}

		image.SetROI(topFaceRight)

		righteye := detector.righteyeCascade.DetectObjects(image)
		var bestEye = 0

		if len(righteye) > 1 {
			bestDist := 0
			dist := 0

			for key, eye := range righteye {
				dist = eye_1.y - (eye.Y() + topFaceRight.Y())
				if dist < bestDist {
					bestEye = key
				}
			}

		}

		eye_2 := pixelCoord{
			x:      righteye[bestEye].X() + topFaceRight.X(),
			y:      righteye[bestEye].Y() + topFaceRight.Y(),
			width:  righteye[bestEye].Width(),
			height: righteye[bestEye].Height(),
		}

		fmt.Println(len(lefteye), " left eyes found")
		fmt.Println(len(righteye), " right eyes found")

		//		if len(lefteye) < 1 || len(righteye) < 1 {
		//			return append(faces, face{
		//				coord:     faceCoords,
		//				eye_left:  pixelCoord{},
		//				eye_right: pixelCoord{},
		//			})
		//		}

		//		eye_2 := pixelCoord{
		//			x:      righteye[0].X() + topFaceRight.X(),
		//			y:      righteye[0].Y() + topFaceRight.Y(),
		//			width:  righteye[0].Width(),
		//			height: righteye[0].Height(),
		//		}

		//
		//		image.ResetROI()
		//		fmt.Println(detFace.X(), int(detFace.Y()+detFace.Height()/2), detFace.Width(), int(detFace.Height()/2))
		//		var lowerFace opencv.Rect
		//		lowerFace.Init(detFace.X(), int(detFace.Y()+detFace.Height()/2), detFace.Width(), int(detFace.Height()/2))
		//		image.SetROI(lowerFace)
		//
		//		mouths := detector.mouthCascade.DetectObjects(image)
		//
		//		fmt.Println(len(mouths), " mouths found")
		//
		//		mouth1 := pixelCoord{
		//			x:      mouths[0].X() + lowerFace.X(),
		//			y:      mouths[0].Y() + lowerFace.Y(),
		//			width:  mouths[0].Width(),
		//			height: mouths[0].Height(),
		//		}

		faces = append(faces, face{
			coord:     faceCoords,
			eye_left:  eye_1,
			eye_right: eye_2,
			//				mouth:     mouth1,
		})

		////		sometimes eyes are inversed ! we switch them
		//				if eye_1.x < eye_2.x {
		//					faces = append(faces, face{
		//						coord:     faceCoords,
		//						eye_left:  eye_1,
		//						eye_right: eye_2,
		//						//				mouth:     mouth1,
		//					})
		//				} else {
		//					faces = append(faces, face{
		//						coord:     faceCoords,
		//						eye_left:  eye_2,
		//						eye_right: eye_1,
		//						//				mouth:     mouth1,
		//					})
		//				}
	}
	fmt.Println(faces)
	return faces
}

func (detector *FaceDetector) Detect(img []byte) faces {

	image := opencv.DecodeImageMem(img)
	faces := detector.DetectFaces(image)

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
