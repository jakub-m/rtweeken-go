package rt

type HittableList []Hittable

var _ (Hittable) = (*HittableList)(nil)

func (list HittableList) Hit(ray Ray, tMin, tMax float64) (bool, HitRecord) {
	tempRec := HitRecord{}
	hitAnything := false
	closestSoFar := tMax

	for _, item := range list {
		if ok, rec := item.Hit(ray, tMin, closestSoFar); ok {
			hitAnything = true
			closestSoFar = rec.T
			tempRec = rec
		}
	}
	return hitAnything, tempRec
}
