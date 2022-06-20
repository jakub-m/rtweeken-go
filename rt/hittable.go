package rt

type Hittable interface {
	Hit(ray Ray, tMin, tMax float64) (bool, HitRecord)
}

type HitRecord struct {
	P         Point3
	T         float64
	Normal    Vec3
	FrontFace bool
	Material  Material
}

func (h HitRecord) SetFaceNormal(ray Ray, outwardNormal Vec3) HitRecord {
	frontFace := ray.Direction.Dot(outwardNormal) < 0
	h.FrontFace = frontFace
	if frontFace {
		h.Normal = outwardNormal
	} else {
		h.Normal = outwardNormal.Neg()
	}
	return h
}
