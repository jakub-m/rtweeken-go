package rt

import "math"

type Material interface {
	Scatter(inRay Ray, rec HitRecord) (bool, Ray, Color)
}

type Lambertian struct {
	Albedo Color
}

func (m Lambertian) Scatter(inRay Ray, rec HitRecord) (bool, Ray, Color) {
	scatterDir := rec.Normal.Add(RandUnitVec3())
	if scatterDir.NearZero() {
		scatterDir = rec.Normal
	}
	scattered := Ray{
		Origin:    rec.P,
		Direction: scatterDir,
	}
	return true, scattered, m.Albedo
}

type Metal struct {
	Albedo Color
	// Fuzz should be in range [0, 1]
	Fuzz float64
}

func (m Metal) Scatter(inRay Ray, rec HitRecord) (bool, Ray, Color) {
	reflected := reflect(UnitVector(inRay.Direction), rec.Normal)
	scattered := Ray{rec.P, reflected.Add(RandVec3InUnitSphere().Mul(m.Fuzz))}
	didScatter := scattered.Direction.Dot(rec.Normal) > 0
	return didScatter, scattered, m.Albedo
}

type Dielectric struct {
	IndexOfRefraction float64
}

func (m Dielectric) Scatter(inRay Ray, rec HitRecord) (bool, Ray, Color) {
	attenuation := NewColor(1.0, 1.0, 1.0)
	var refractionRatio float64
	if rec.FrontFace {
		refractionRatio = (1.0 / m.IndexOfRefraction)
	} else {
		refractionRatio = m.IndexOfRefraction
	}
	unitDirection := UnitVector(inRay.Direction)
	cosTheta := math.Min(unitDirection.Neg().Dot(rec.Normal), 1.0)
	sinTheta := math.Sqrt(1.0 - cosTheta*cosTheta)
	cannotRefract := refractionRatio*sinTheta > 1.0
	var direction Vec3
	if cannotRefract || reflectance(cosTheta, refractionRatio) > RandFloat() {
		direction = reflect(unitDirection, rec.Normal)
	} else {
		direction = refract(unitDirection, rec.Normal, refractionRatio)
	}
	scattered := Ray{rec.P, direction}
	return true, scattered, attenuation
}

func reflect(v, n Vec3) Vec3 {
	return v.Sub(n.Mul(2 * v.Dot(n)))
}

func refract(uv, n Vec3, etaiOverEtat float64) Vec3 {
	cosTheta := math.Min(uv.Neg().Dot(n), 1.0)
	rOutPerp := (uv.Add(n.Mul(cosTheta))).Mul(etaiOverEtat)
	rOutParallel := n.Mul(-math.Sqrt(math.Abs(1.0 - rOutPerp.LengthSquared())))
	return rOutPerp.Add(rOutParallel)
}

func reflectance(cosine, refIdx float64) float64 {
	// Use Schlick's approximation for reflectance.
	r0 := (1 - refIdx) / (1 + refIdx)
	r0 = r0 * r0
	return r0 + (1-r0)*math.Pow((1-cosine), 5)
}
