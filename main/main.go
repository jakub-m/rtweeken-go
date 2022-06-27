package main

import (
	"flag"
	"fmt"
	"image"
	gocolor "image/color"
	"image/color/palette"
	"image/gif"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"

	"raytracing/rt"
)

const (
	rowsPerWorker = 10
)

type sceneDef interface {
	getCamera(radians float64) rt.Camera
	getWorld() rt.Hittable
}

var scenes = map[string]sceneDef{
	"simple": simpleScene{},
	"balls":  randomScene{},
}

func main() {
	rand.Seed(0)
	opts := renderOpts{}
	var outputFile string
	var animate bool
	var numFrames int
	var sceneName string
	var angleRadians float64

	flag.StringVar(&outputFile, "o", "scene.png", "output file name")
	flag.IntVar(&opts.samplesPerPixel, "p", 1, "samples per pixel")
	flag.IntVar(&opts.maxDepth, "d", 50, "maximum depth, number of ray reflections")
	flag.IntVar(&opts.width, "w", 320, "image width")
	flag.Float64Var(&angleRadians, "a", 30, "camera angle degrees")
	flag.BoolVar(&animate, "animate", false, "gif animation")
	flag.IntVar(&numFrames, "frames", 30, "number of animation frames")
	flag.StringVar(&sceneName, "scene", "simple", fmt.Sprintf("scene to render (%s)", strings.Join(keys(scenes), ",")))
	flag.IntVar(&opts.workerCount, "n", runtime.NumCPU(), "worker count")
	flag.Parse()

	if animate && outputFile == "scene.png" {
		outputFile = "scene.gif"
	}

	if animate && !strings.HasSuffix(outputFile, ".gif") {
		log.Fatal("use .gif file for animation")
	}
	if !animate && !strings.HasSuffix(outputFile, ".png") {
		log.Fatal("use .png file for scene")
	}

	opts.height = int(float32(opts.width) / rt.AspectRatio)
	angleRadians = rt.DegreesToRadians(angleRadians)

	var sceneDef sceneDef
	if f, ok := scenes[sceneName]; ok {
		sceneDef = f
	} else {
		log.Fatalf("bad scene name: %s", sceneName)
	}

	out, err := os.Create(outputFile)
	rt.CheckNoError(err)
	defer out.Close()

	if animate {
		gifImage := &gif.GIF{
			Image: []*image.Paletted{},
		}

		world := sceneDef.getWorld()
		for i := 0; i < numFrames; i++ {
			fmt.Printf("frame %d/%d\n", i+1, numFrames)
			actualAngle := angleRadians + (2.0 * math.Pi / float64(numFrames) * float64(i))
			camera := sceneDef.getCamera(actualAngle)
			frame := renderImage(opts, world, camera)
			gifFrame := imageToGif(frame)
			gifImage.Image = append(gifImage.Image, gifFrame)
			gifImage.Delay = append(gifImage.Delay, 1)
		}
		err = gif.EncodeAll(out, gifImage)
		rt.CheckNoError(err)
	} else {
		world := sceneDef.getWorld()
		camera := sceneDef.getCamera(angleRadians)
		image := renderImage(opts, world, camera)
		err = png.Encode(out, image)
		rt.CheckNoError(err)
	}
}

type renderOpts struct {
	samplesPerPixel int
	maxDepth        int
	width           int
	height          int
	camera          rt.Camera
	world           rt.Hittable
	workerCount     int
}

type simpleScene struct{}

func (s simpleScene) getWorld() rt.Hittable {
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
	return world
}

func (s simpleScene) getCamera(angleRadians float64) rt.Camera {
	x, z := coordsOnRing(4, angleRadians)
	lookFrom := rt.NewPoint3(x, 2, z)
	lookAt := rt.NewPoint3(0, 0, -1)
	focusDist := (lookFrom.Sub(lookAt.Vec3)).Length()
	cam := rt.NewCamera(
		lookFrom, lookAt,
		20, 0.1, focusDist)
	return cam
}

type randomScene struct{}

func (s randomScene) getWorld() rt.Hittable {
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

	return world
}

func (s randomScene) getCamera(angleRadians float64) rt.Camera {
	x, z := coordsOnRing(13, angleRadians)
	// lookFrom := rt.NewPoint3(13, 2, 3)
	lookFrom := rt.NewPoint3(x, 2, z)
	lookAt := rt.NewPoint3(0, 0, 0)
	focusDist := 10.0
	aperture := 0.10
	vfov := 30.0
	cam := rt.NewCamera(lookFrom, lookAt, vfov, aperture, focusDist)
	return cam
}

func renderImage(opts renderOpts, world rt.Hittable, cam rt.Camera) *image.RGBA {
	wg := new(sync.WaitGroup)
	rowNumberChan := make(chan []int)
	rowResultChan := make(chan []pixelRGBA)
	wg.Add(opts.workerCount)

	go func() {
		rowPack := []int{}
		for j := 0; j < opts.height; j++ {
			rowPack = append(rowPack, j)
			// If we not bundle rows per worker, then the overhead of synchronization
			// will eat the large part of parallel processing.
			if len(rowPack) == rowsPerWorker {
				rowNumberChan <- rowPack
				rowPack = []int{}
			}
		}
		rowNumberChan <- rowPack
		close(rowNumberChan)
	}()
	for i := 0; i < opts.workerCount; i++ {
		go lineRenderingWorker(i+1, wg, rowNumberChan, rowResultChan, opts, world, cam)
	}
	go func() {
		wg.Wait()
		close(rowResultChan)
	}()

	image := image.NewRGBA(image.Rect(0, 0, opts.width, opts.height))
	for renderedLine := range rowResultChan {
		for _, p := range renderedLine {
			image.Set(p.x, p.y, p.c)
		}
	}
	return image
}

func lineRenderingWorker(
	workerId int,
	wg *sync.WaitGroup,
	rowNumChan <-chan []int,
	rowResultChan chan<- []pixelRGBA,
	opts renderOpts,
	world rt.Hittable,
	cam rt.Camera,
) {
	for rowNumbers := range rowNumChan {
		for _, j := range rowNumbers {
			log.Printf("worker %d line %d\n", workerId, j)
			renderedLine := renderLine(j, opts, world, cam)
			rowResultChan <- renderedLine
		}
	}
	wg.Done()
}

type pixelRGBA struct {
	x, y int
	c    gocolor.RGBA
}

func renderLine(j int, opts renderOpts, world rt.Hittable, cam rt.Camera) []pixelRGBA {
	rendered := []pixelRGBA{}
	for i := 0; i < opts.width; i++ {
		pixelColor := rt.NewColor(0, 0, 0)
		for s := 0; s < opts.samplesPerPixel; s++ {
			u := (float64(i) + rt.RandFloat()) / float64(opts.width-1)
			v := (float64(j) + rt.RandFloat()) / float64(opts.height-1)
			ray := cam.GetRay(u, v)
			rc := rt.RayColor(ray, world, opts.maxDepth)
			pixelColor = rt.Color{pixelColor.Add(rc.Vec3)}
		}
		x := i
		y := (opts.height - 1 - j)
		color := getColor(opts, x, y, pixelColor)
		rendered = append(rendered, pixelRGBA{x, y, color})
	}
	return rendered
}

func getColor(opts renderOpts, x, y int, c rt.Color) gocolor.RGBA {
	scale := 1.0 / float64(opts.samplesPerPixel)
	// sqrt is for Gamma 2 correction
	r := math.Sqrt(c.R() * scale)
	g := math.Sqrt(c.G() * scale)
	b := math.Sqrt(c.B() * scale)
	return gocolor.RGBA{
		uint8(256 * clamp(r, 0.0, 0.999)),
		uint8(256 * clamp(g, 0.0, 0.999)),
		uint8(256 * clamp(b, 0.0, 0.999)),
		255,
	}
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

func coordsOnRing(radius, angleRad float64) (float64, float64) {
	x := math.Cos(angleRad) * radius
	y := math.Sin(angleRad) * radius
	return x, y
}

func imageToGif(orig *image.RGBA) *image.Paletted {
	paletted := image.NewPaletted(orig.Rect, palette.Plan9)
	for y := orig.Rect.Min.Y; y < orig.Rect.Max.Y; y++ {
		for x := orig.Rect.Min.X; x < orig.Rect.Max.X; x++ {
			paletted.SetRGBA64(x, y, orig.RGBA64At(x, y))
		}
	}
	return paletted
}

func keys[T any](m map[string]T) []string {
	kk := []string{}
	for k := range m {
		kk = append(kk, k)
	}
	return kk
}
