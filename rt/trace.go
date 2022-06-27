package rt

import "math"

func RayColor(r Ray, world Hittable, depth int) Color {
	if depth <= 0 {
		return NewColor(0, 0, 0)
	}

	// 0.001 is to fix the "shadow acne problem".
	if ok, rec := world.Hit(r, 0.001, math.MaxFloat64); ok {
		if wasScattered, scattered, attenuation := rec.Material.Scatter(r, rec); wasScattered {
			return Color{attenuation.Vec3.MulVec(RayColor(scattered, world, depth-1).Vec3)}
		}
		return NewColor(0, 0, 0)
	}

	unitDirection := UnitVector(r.Direction)
	t := 0.5 * (unitDirection.Y() + 1.0)
	return Color{
		Vec3: NewColor(1, 1, 1).Mul(1 - t).Add(NewColor(0.5, 0.7, 1.0).Mul(t)),
	}
}
