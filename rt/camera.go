package rt

import "math"

const (
	AspectRatio = 16.0 / 9.0
	// viewportHeight = 2
	// viewportWidth  = viewportHeight * AspectRatio
	focalLength = 1
)

type Camera struct {
	origin          Point3
	horizontal      Vec3
	vertical        Vec3
	lowerLeftCorner Vec3
	lensRadius      float64
	u, v            Vec3
}

func NewCamera(
	lookFrom, lookAt Point3,
	vertFovDeg, aperture, focusDist float64,
) Camera {
	theta := degreesToRadians(vertFovDeg)
	h := math.Tan(theta / 2)
	viewportHeight := 2 * h
	viewportWidth := AspectRatio * viewportHeight

	vup := Vec3{0, 1, 0}
	w := UnitVector(lookFrom.Sub(lookAt.Vec3))
	u := UnitVector(vup.Cross(w))
	v := w.Cross(u)

	origin := lookFrom
	// here
	horizontal := u.Mul(viewportWidth * focusDist)
	vertical := v.Mul(viewportHeight * focusDist)
	lowerLeftCorner := origin.Sub(horizontal.Div(2)).Sub(vertical.Div(2)).Sub(w.Mul(focusDist))
	lensRadius := aperture / 2
	return Camera{
		origin:          origin,
		horizontal:      horizontal,
		vertical:        vertical,
		lowerLeftCorner: lowerLeftCorner,
		lensRadius:      lensRadius,
		u:               u,
		v:               v,
	}
}

func (c Camera) GetRay(s, t float64) Ray {
	rd := randomInUnitDisk().Mul(c.lensRadius)
	offset := c.u.Mul(rd.X()).Add(c.v.Mul(rd.Y()))
	return Ray{
		Origin:    Point3{c.origin.Add(offset)},
		Direction: c.lowerLeftCorner.Add(c.horizontal.Mul(s)).Add(c.vertical.Mul(t)).Sub(c.origin.Vec3).Sub(offset),
	}
}

func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

func randomInUnitDisk() Vec3 {
	for {
		p := Vec3{RandFloatRange(-1, 1), RandFloatRange(-1, 1), 0}
		if p.LengthSquared() < 1 {
			return p
		}
	}
}
