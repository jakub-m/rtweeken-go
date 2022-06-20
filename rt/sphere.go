package rt

import "math"

type Sphere struct {
	Center  Point3
	Radius  float64
	Matrial Material
}

var _ (Hittable) = (*Sphere)(nil)

func NewSphere(center Point3, radius float64, material Material) Sphere {
	return Sphere{center, radius, material}
}

func (s Sphere) Hit(ray Ray, tMin, tMax float64) (bool, HitRecord) {
	oc := ray.Origin.Sub(s.Center.Vec3)
	a := ray.Direction.LengthSquared()
	halfB := oc.Dot(ray.Direction)
	c := oc.LengthSquared() - s.Radius*s.Radius
	discriminant := halfB*halfB - a*c
	if discriminant < 0 {
		return false, HitRecord{}
	}

	sqrtd := math.Sqrt(discriminant)
	// Find the nearest root that lies in the acceptable range.
	root := (-halfB - sqrtd) / a
	if root < tMin || tMax < root {
		root = (-halfB + sqrtd) / a
		if root < tMin || tMax < root {
			return false, HitRecord{}
		}
	}

	recT := root
	recP := ray.At(recT)
	rec := HitRecord{
		P:        recP,
		T:        recT,
		Material: s.Matrial,
	}
	outwardNormal := (recP.Sub(s.Center.Vec3)).Div(s.Radius)
	rec = rec.SetFaceNormal(ray, outwardNormal)
	return true, rec
}
