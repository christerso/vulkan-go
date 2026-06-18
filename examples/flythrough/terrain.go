package main

import "math"

// Vertex is a terrain vertex: position and normal. 24 bytes, POD.
type Vertex struct {
	Px, Py, Pz float32
	Nx, Ny, Nz float32
}

// Terrain holds a generated heightmap mesh.
type Terrain struct {
	Vertices    []Vertex
	Indices     []uint32
	HeightScale float32
	WorldSize   float32
}

// hash returns a deterministic pseudo-random value in [0,1) for a grid cell.
func hash(x, y int) float32 {
	h := uint32(x)*374761393 + uint32(y)*668265263
	h = (h ^ (h >> 13)) * 1274126177
	h ^= h >> 16
	return float32(h) / float32(1<<32)
}

// smooth value-noise interpolation.
func smooth(a, b, t float32) float32 {
	t = t * t * (3 - 2*t)
	return a + (b-a)*t
}

func valueNoise(x, y float32) float32 {
	xi, yi := int(math.Floor(float64(x))), int(math.Floor(float64(y)))
	xf, yf := x-float32(xi), y-float32(yi)
	v00, v10 := hash(xi, yi), hash(xi+1, yi)
	v01, v11 := hash(xi, yi+1), hash(xi+1, yi+1)
	return smooth(smooth(v00, v10, xf), smooth(v01, v11, xf), yf)
}

// fbm sums octaves of value noise.
func fbm(x, y float32) float32 {
	var sum, amp, freq float32 = 0, 0.5, 1
	for i := 0; i < 6; i++ {
		sum += amp * valueNoise(x*freq, y*freq)
		freq *= 2
		amp *= 0.5
	}
	return sum
}

// GenerateTerrain builds an n-by-n vertex grid spanning worldSize units with
// fbm height. Normals come from height finite differences.
func GenerateTerrain(n int, worldSize, heightScale float32) Terrain {
	if n < 2 {
		n = 2
	}
	verts := make([]Vertex, n*n)
	step := worldSize / float32(n-1)
	const features = 6.0 // number of large hills across the map

	height := func(ix, iy int) float32 {
		if ix < 0 {
			ix = 0
		} else if ix >= n {
			ix = n - 1
		}
		if iy < 0 {
			iy = 0
		} else if iy >= n {
			iy = n - 1
		}
		fx := float32(ix) / float32(n-1) * features
		fy := float32(iy) / float32(n-1) * features
		h := fbm(fx, fy)
		// sharpen valleys for a mountain feel
		h = h * h
		return (h - 0.3) * heightScale
	}

	for iy := 0; iy < n; iy++ {
		for ix := 0; ix < n; ix++ {
			i := iy*n + ix
			px := float32(ix)*step - worldSize*0.5
			pz := float32(iy)*step - worldSize*0.5
			py := height(ix, iy)
			// normal from neighbor heights
			hl := height(ix-1, iy)
			hr := height(ix+1, iy)
			hd := height(ix, iy-1)
			hu := height(ix, iy+1)
			nx := hl - hr
			nz := hd - hu
			ny := 2 * step
			inv := float32(1.0) / float32(math.Sqrt(float64(nx*nx+ny*ny+nz*nz)))
			verts[i] = Vertex{
				Px: px, Py: py, Pz: pz,
				Nx: nx * inv, Ny: ny * inv, Nz: nz * inv,
			}
		}
	}

	indices := make([]uint32, 0, (n-1)*(n-1)*6)
	for iy := 0; iy < n-1; iy++ {
		for ix := 0; ix < n-1; ix++ {
			i0 := uint32(iy*n + ix)
			i1 := i0 + 1
			i2 := i0 + uint32(n)
			i3 := i2 + 1
			indices = append(indices, i0, i2, i1, i1, i2, i3)
		}
	}

	return Terrain{Vertices: verts, Indices: indices, HeightScale: heightScale, WorldSize: worldSize}
}
