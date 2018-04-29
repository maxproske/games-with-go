package vec3

import "math"

// Publically visible
type Vector3 struct {
	X, Y, Z float32
}

// Linear algebra in disguise as something fun.
// A vector is a point (x, y) which computes a magnitude and length
func main() {

}

func Add(a, b Vector3) Vector3 {
	return Vector3{a.X + b.X, a.Y + b.Y, a.Z + b.Z}
}

func Mult(a Vector3, b float32) Vector3 {
	return Vector3{a.X * b, a.Y * b, a.Z * b}
}

// Not a pointer, because 3 float32's is not a lot to copy.
func (a Vector3) Length() float32 {
	return float32(math.Sqrt(float64(a.X*a.X + a.Y*a.Y + a.Z*a.Z)))
}

// Calculate distnace between vectors
func Distance(a, b Vector3) float32 {
	xDiff := a.X - b.X
	yDiff := a.Y - b.Y
	zDiff := a.Z - b.Z
	return float32(math.Sqrt(float64(xDiff*xDiff + yDiff*yDiff + zDiff*zDiff)))
}

// Sometimes we just need to know if one thing is close to another, and Sqrt is an expensive operation.
func DistanceSquared(a, b Vector3) float32 {
	xDiff := a.X - b.X
	yDiff := a.Y - b.Y
	zDiff := a.Z - b.Z
	return xDiff*xDiff + yDiff*yDiff + zDiff*zDiff
}

// Resize length to 1, if we just want to know the direction
func Normalize(a Vector3) Vector3 {
	len := a.Length()
	return Vector3{a.X / len, a.Y / len, a.Z / len}
}
