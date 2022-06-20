// Go strengths:
// - Renaming variables or types is a breeze.
// - Fast iterations. It either did not compile or it worked.
// Go quirks:
// - I couldn't put all the code in a single "main" package with many files within the "package main".
//   I had to have a namespaced module with all the library code and the main.
// Go drawbacks:
// - Lack of implicit type (un)aliasing makes the syntach cumersome. See the final vec3.
// - Warning on unkeyed fields of composite literal, when I didn't want to use composite literals.
//   I was forced to use the composite literal to be able to access the shared Vec3 methods.
// - Mathematical expressions that otherwise feel natural in python or C++ look awkward in Go.
//   This is because we cannot overload math operators. Compare
//   Original C++: v - 2*dot(v,n)*n;
//   Go: v.Sub(n.Mul(2 * v.Dot(n)))

package main

import (
	"image"
	gocolor "image/color"
	"image/png"
	"log"
	"math"
	"os"

	"raytracing/rt"
)

const (
	samplesPerPixel = 300
	maxDepth        = 50
	outputFile      = "scene.png"
	imageWidth      = 640 // TODO how to convert those fractional costants to ints? 80, 320
	imageHeight     = imageWidth / rt.AspectRatio
)

func main() {
	out, err := os.Create(outputFile)
	rt.CheckNoError(err)
	defer out.Close()

	//image := renderTestScene()
	image := renderRandomScene()

	err = png.Encode(out, image)
	rt.CheckNoError(err)
}

func renderTestScene() *image.RGBA {
	image := image.NewRGBA(image.Rect(0, 0, int(imageWidth), int(imageHeight)))

	materialGround := rt.Lambertian{
		Albedo: rt.NewColor(0.8, 0.8, 0.0),
	}
	materialCenter := rt.Lambertian{
		Albedo: rt.NewColor(0.1, 0.2, 0.5),
	}
	materialLeft := rt.Dielectric{
		IndexOfRefraction: 1.5,
	}
	materialRight := rt.Metal{
		Albedo: rt.NewColor(0.8, 0.6, 0.2),
		Fuzz:   0.0,
	}

	world := rt.HittableList{
		rt.NewSphere(rt.NewPoint3(0, -100.5, -1), 100, materialGround),
		rt.NewSphere(rt.NewPoint3(0, 0, -1), 0.5, materialCenter),
		rt.NewSphere(rt.NewPoint3(-1.0, 0.0, -1.0), 0.5, materialLeft),
		rt.NewSphere(rt.NewPoint3(-1.0, 0.0, -1.0), -0.45, materialLeft),
		rt.NewSphere(rt.NewPoint3(1.0, 0.0, -1.0), 0.5, materialRight),
	}

	lookFrom := rt.NewPoint3(-2, 2, 1)
	lookAt := rt.NewPoint3(0, 0, -1)
	focusDist := (lookFrom.Sub(lookAt.Vec3)).Length()
	cam := rt.NewCamera(
		lookFrom, lookAt,
		20, 2.0, focusDist)

	renderImage(image, cam, world)

	return image
}

func renderRandomScene() *image.RGBA {
	image := image.NewRGBA(image.Rect(0, 0, int(imageWidth), int(imageHeight)))

	world := rt.HittableList{}
	ground_material := rt.Lambertian{
		Albedo: rt.NewColor(0.5, 0.5, 0.5),
	}
	world = append(world, rt.NewSphere(rt.NewPoint3(0, -1000, 0), 1000, ground_material))

	for a := -11; a < 11; a++ {
		for b := -11; b < 11; b++ {
			choose_mat := rt.RandFloat()
			center := rt.NewPoint3(float64(a)+0.9*rt.RandFloat(), 0.2, float64(b)+0.9*rt.RandFloat())

			if (center.Sub(rt.Vec3{4.0, 0.2, 0.0}).Length()) > 0.9 {
				if choose_mat < 0.8 {
					// diffuse
					albedo := rt.Color{rt.RandVec3(0, 1).MulVec(rt.RandVec3(0, 1))}
					sphere_material := rt.Lambertian{Albedo: albedo}
					world = append(world, rt.NewSphere(center, 0.2, sphere_material))
				} else if choose_mat < 0.95 {
					// metal
					albedo := rt.Color{rt.RandVec3(0.5, 1)}
					fuzz := rt.RandFloatRange(0.0, 0.5)
					sphere_material := rt.Metal{albedo, fuzz}
					world = append(world, rt.NewSphere(center, 0.2, sphere_material))
				} else {
					// glass
					sphere_material := rt.Dielectric{1.5}
					world = append(world, rt.NewSphere(center, 0.2, sphere_material))
				}
			}
		}
	}

	material1 := rt.Dielectric{1.5}
	world = append(world, rt.NewSphere(rt.NewPoint3(0, 1, 0), 1.0, material1))

	material2 := rt.Lambertian{rt.NewColor(0.4, 0.2, 0.1)}
	world = append(world, rt.NewSphere(rt.NewPoint3(-4, 1, 0), 1.0, material2))

	material3 := rt.Metal{rt.NewColor(0.7, 0.6, 0.5), 0.0}
	world = append(world, rt.NewSphere(rt.NewPoint3(4, 1, 0), 1.0, material3))

	lookFrom := rt.NewPoint3(13, 2, 3)
	lookAt := rt.NewPoint3(0, 0, 0)
	focusDist := 10.0
	aperture := 0.10
	vfov := 30.0
	cam := rt.NewCamera(lookFrom, lookAt, vfov, aperture, focusDist)

	renderImage(image, cam, world)

	return image
}

func renderImage(image *image.RGBA, cam rt.Camera, world rt.Hittable) {
	for j := int(imageHeight); j >= 0; j-- { // The original had --j and ++i
		log.Println(j)
		for i := 0; i < int(imageWidth); i++ {
			pixelColor := rt.NewColor(0, 0, 0)
			for s := 0; s < samplesPerPixel; s++ {
				u := (float64(i) + rt.RandFloat()) / (imageWidth - 1)
				v := (float64(j) + rt.RandFloat()) / (imageHeight - 1)
				ray := cam.GetRay(u, v)
				rc := rt.RayColor(ray, world, maxDepth)
				pixelColor = rt.Color{pixelColor.Add(rc.Vec3)}
			}
			writeColor(image, i, (imageHeight - 1 - j), pixelColor)
		}
	}
}

func writeColor(img *image.RGBA, x, y int, c rt.Color) {
	scale := 1.0 / samplesPerPixel
	// sqrt is for Gamma 2 correction
	r := math.Sqrt(c.R() * scale)
	g := math.Sqrt(c.G() * scale)
	b := math.Sqrt(c.B() * scale)
	img.Set(x, y, gocolor.RGBA{
		uint8(256 * clamp(r, 0.0, 0.999)),
		uint8(256 * clamp(g, 0.0, 0.999)),
		uint8(256 * clamp(b, 0.0, 0.999)),
		255,
	})
}

func randomInHemisphere(normal rt.Vec3) rt.Vec3 {
	us := rt.RandVec3InUnitSphere()
	if us.Dot(normal) > 0.0 { // In the same hemisphere as the normal
		return us
	} else {
		return us.Neg()
	}
}

func clamp(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}
