package rt

import "math"

// Go drawback: for type aliases, the functions of the base (aliased) type are not available to the alias. One needs to un-alias to
// base type to apply the methods, for example:
//
// p := point3{0, 0, 0}
// vec3(p).sub(vec3(p))
//
// One needs to express aliases as composites
//
// type foo struct {
//     baseType
// }

// Go drawback (which might be considered a feature): lack of possibility of overloading infix and other built-in operators makes it
// cumbersome to express math operations, line foo + bar instead of foo.Add(bar)

type Point3 struct {
	Vec3
}

func NewPoint3(x, y, z float64) Point3 {
	return Point3{Vec3{x, y, z}}
}

type Color struct {
	Vec3
}

func NewColor(r, g, b float64) Color {
	return Color{Vec3{r, g, b}}
}

func (c Color) R() float64 {
	return c.Vec3[0]
}

func (c Color) G() float64 {
	return c.Vec3[1]
}

func (c Color) B() float64 {
	return c.Vec3[2]
}

type Vec3 [3]float64

func (v Vec3) X() float64 {
	return v[0]
}

func (v Vec3) Y() float64 {
	return v[1]
}

func (v Vec3) Z() float64 {
	return v[2]
}

func (v Vec3) Neg() Vec3 {
	v[0] *= -1
	v[1] *= -1
	v[2] *= -1
	return v
}

func (v Vec3) Add(o Vec3) Vec3 {
	v[0] += o[0]
	v[1] += o[1]
	v[2] += o[2]
	return v
}

func (v Vec3) Sub(o Vec3) Vec3 {
	v[0] -= o[0]
	v[1] -= o[1]
	v[2] -= o[2]
	return v
}

func (v Vec3) Mul(c float64) Vec3 {
	v[0] *= c
	v[1] *= c
	v[2] *= c
	return v
}

func (v Vec3) Div(c float64) Vec3 {
	v[0] /= c
	v[1] /= c
	v[2] /= c
	return v
}

func (v Vec3) MulVec(o Vec3) Vec3 {
	v[0] *= o[0]
	v[1] *= o[1]
	v[2] *= o[2]
	return v
}

func (v Vec3) Dot(o Vec3) float64 {
	return v[0]*o[0] + v[1]*o[1] + v[2]*o[2]
}

func (v Vec3) Cross(o Vec3) Vec3 {
	return Vec3{
		v[1]*o[2] - v[2]*o[1],
		v[2]*o[0] - v[0]*o[2],
		v[0]*o[1] - v[1]*o[0],
	}
}

func (v Vec3) Length() float64 {
	return math.Sqrt(v.LengthSquared())
}

func (v Vec3) LengthSquared() float64 {
	return v.Dot(v)
}

func (v Vec3) NearZero() bool {
	s := 1e-8
	abs := math.Abs
	return (abs(v[0]) < s) && (abs(v[1]) < s) && (abs(v[2]) < s)
}

func RandVec3(min, max float64) Vec3 {
	return Vec3{
		RandFloatRange(min, max),
		RandFloatRange(min, max),
		RandFloatRange(min, max),
	}
}

func RandUnitVec3() Vec3 {
	return UnitVector(RandVec3InUnitSphere())
}

func UnitVector(v Vec3) Vec3 {
	return v.Div(v.Length())
}

func RandVec3InUnitSphere() Vec3 {
	for {
		v := RandVec3(-1, 1)
		if v.LengthSquared() < 1 {
			return v
		}
	}
}
