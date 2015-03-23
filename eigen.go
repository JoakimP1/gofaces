package gofaces

import (
	gomat "github.com/skelterjohn/go.matrix"
	"math"
)

type EigenFace struct {
	rowCount int
	colCount int

	pixelMatrix [][]float64
	meanMatrix  []float64
	diffMatrix  [][]float64
	covMatrix   [][]float64
	eigenMatrix [][]float64
}

func NewMatrix(rows, cols int) [][]float64 {
	matrix := make([][]float64, rows, rows)
	for i := 0; i < rows; i++ {
		matrix[i] = make([]float64, cols, cols)
	}
	return matrix
}

func NewEigenFace(width int, height int, pixels [][]byte) *EigenFace {

	rows := len(pixels)
	cols := width * height

	eigenFace := EigenFace{
		rowCount:    rows,
		colCount:    cols,
		pixelMatrix: NewMatrix(rows, cols),
		meanMatrix:  make([]float64, rows, cols),
		diffMatrix:  NewMatrix(rows, cols),
		covMatrix:   NewMatrix(cols, cols),
		eigenMatrix: NewMatrix(rows, cols),
	}

	return &eigenFace
}

func (eigen *EigenFace) Train() {

	eigen.computeMeanColumn()
	eigen.computeDifferenceMatrixPixels()
	eigen.computeCovarianceMatrix()
	eigen.computeEigenFaces()
}

func (eigen *EigenFace) computeMeanColumn() {

	for k := 0; k < eigen.rowCount; k++ {
		sum := 0.0
		for l := 0; l < eigen.colCount; l++ {
			sum += eigen.pixelMatrix[k][l]
		}
		eigen.meanMatrix[k] = sum / float64(eigen.colCount)
	}

}

func (eigen *EigenFace) computeDifferenceMatrixPixels() {
	for i := 0; i < eigen.rowCount; i++ {
		for j := 0; j < eigen.colCount; j++ {
			eigen.diffMatrix[i][j] = eigen.pixelMatrix[i][j] - eigen.meanMatrix[i]
		}
	}
}

func (eigen *EigenFace) computeCovarianceMatrix() {
	for i := 0; i < eigen.colCount; i++ {
		for j := 0; j < eigen.colCount; j++ {
			sum := 0.0
			for k := 0; k < eigen.rowCount; k++ {
				sum += eigen.diffMatrix[k][i] * eigen.diffMatrix[k][j]
			}
			eigen.covMatrix[i][j] = sum
		}
	}
}

func (eigen *EigenFace) ComputeDistance(subjectPixels []float64) float64 {
	diffPixels := eigen.ComputeDifferencePixels(subjectPixels)
	weights := eigen.ComputeWeights(diffPixels)
	reconstructedEigenPixels := eigen.ReconstructImageWithEigenFaces(weights)
	return eigen.ComputeImageDistance(subjectPixels, reconstructedEigenPixels)

}

func (eigen *EigenFace) ComputeImageDistance(pixels1, pixels2 []float64) float64 {

	distance := 0.0
	for i := 0; i < eigen.rowCount; i++ {
		diff := pixels1[i] - pixels2[i]
		distance += diff * diff
	}

	return math.Sqrt(distance / float64(eigen.rowCount))
}

func (eigen *EigenFace) ComputeDifferencePixels(subjectPixels []float64) []float64 {
	diffPixels := make([]float64, eigen.rowCount, eigen.rowCount)
	for i := 0; i < eigen.rowCount; i++ {
		diffPixels[i] = subjectPixels[i] - eigen.meanMatrix[i]
	}
	return diffPixels
}

func (eigen *EigenFace) ComputeWeights(diffImagePixels []float64) []float64 {
	eigenWeights := make([]float64, eigen.rowCount, eigen.rowCount)
	for i := 0; i < eigen.colCount; i++ {
		for j := 0; j < eigen.rowCount; j++ {
			eigenWeights[i] += diffImagePixels[j] * eigen.eigenMatrix[j][i]
		}
	}

	return eigenWeights
}

func (eigen *EigenFace) ReconstructImageWithEigenFaces(weights []float64) []float64 {
	reconstructedPixels := make([]float64, eigen.rowCount, eigen.rowCount)

	for i := 0; i < eigen.colCount; i++ {
		for j := 0; j < eigen.rowCount; j++ {
			reconstructedPixels[j] += weights[i] * eigen.eigenMatrix[j][i]
		}
	}

	for i := 0; i < eigen.rowCount; i++ {
		reconstructedPixels[i] += eigen.meanMatrix[i]
	}

	min := float64(math.MaxFloat64)
	max := float64(-math.MaxFloat64)

	for i := 0; i < eigen.rowCount; i++ {
		min = math.Min(min, reconstructedPixels[i])
		max = math.Max(max, reconstructedPixels[i])
	}

	normalizedReconstructedPixels := make([]float64, eigen.rowCount, eigen.rowCount)
	for i := 0; i < eigen.rowCount; i++ {
		normalizedReconstructedPixels[i] = (255.0 * (reconstructedPixels[i] - min)) / (max - min)
	}

	return normalizedReconstructedPixels
}

func (eigen *EigenFace) computeEigenFaces() {

	denseMat := gomat.MakeDenseMatrixStacked(eigen.covMatrix)
	eigenVectors, _, _ := denseMat.Eigen()

	imageCount := eigenVectors.Cols()
	rank := eigenVectors.Rows()

	for i := 0; i < rank; i++ {
		sumSquare := 0.0
		for j := 0; j < eigen.rowCount; j++ {
			for k := 0; k < imageCount; k++ {

				eigen.eigenMatrix[j][i] += eigen.diffMatrix[j][k] * eigenVectors.Get(i, k)
			}
			sumSquare += eigen.eigenMatrix[j][i] * eigen.eigenMatrix[j][i]
		}
		norm := math.Sqrt(float64(sumSquare))
		for j := 0; j < eigen.rowCount; j++ {
			eigen.eigenMatrix[j][i] /= norm
		}
	}

}
