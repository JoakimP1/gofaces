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
}

type faces []face

func (face *face) Angle() float64 {

	r := math.Atan2(float64(face.eye_right.y-face.eye_left.y), float64(face.eye_right.x-face.eye_left.x))
	if r > 0.0 {
		return r
	} else {
		//		return (-r * 360) / math.Pi
		return (-r * 180) / math.Pi
	}
}

var eyeCascade *opencv.HaarCascade
var faceCascade *opencv.HaarCascade

func PaintFace(img []byte, face face) []byte {

	image := opencv.DecodeImageMem(img)

	opencv.Rectangle(image,
		opencv.Point{face.coord.x + face.coord.width, face.coord.y},
		opencv.Point{face.coord.x, face.coord.y + face.coord.height},
		opencv.ScalarAll(0), 1, 1, 0)

	opencv.Circle(image,
		opencv.Point{face.eye_left.x + face.eye_left.width/2, face.eye_left.y + face.eye_left.height/2},
		2,
		opencv.ScalarAll(255), 1, 1, 0)

	opencv.Circle(image,
		opencv.Point{face.eye_right.x + face.eye_right.width/2, face.eye_right.y + face.eye_right.height/2},
		2,
		opencv.ScalarAll(255), 1, 1, 0)

	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, image.ToImage(), nil)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func Detect(img []byte) faces {
	eyeCascade = opencv.LoadHaarClassifierCascade("/home/joakim/opencv-2.4.9/data/haarcascades/haarcascade_eye.xml")

	faceCascade = opencv.LoadHaarClassifierCascade("/home/joakim/Go/src/github.com/lazywei/go-opencv/samples/haarcascade_frontalface_alt.xml")
	image := opencv.DecodeImageMem(img)

	var faces = make(faces, 0)

	detFaces := faceCascade.DetectObjects(image)
	if len(detFaces) != 1 {
		if len(detFaces) > 1 {
			panic("To many faces in image")
		} else {
			panic("No faces found in image")
		}
	}

	for _, detFace := range detFaces {

		faceCoords := pixelCoord{
			x:      detFace.X(),
			y:      detFace.Y(),
			width:  detFace.Width(),
			height: detFace.Height(),
		}

		eyes := eyeCascade.DetectObjects(image)
		if len(eyes) < 2 {
			fmt.Println(len(eyes), " eyes found")
			return append(faces, face{
				coord:     faceCoords,
				eye_left:  pixelCoord{},
				eye_right: pixelCoord{},
			})
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
	}

	return faces
}
