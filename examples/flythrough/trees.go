package main

import "math"

// treeVertex is a tree mesh vertex with a baked base color.
type treeVertex struct {
	px, py, pz float32
	nx, ny, nz float32
	cr, cg, cb float32
}

// treeInstance is per-tree placement data.
type treeInstance struct {
	ox, oy, oz float32
	tr, tg, tb float32
	scale      float32
}

// treeMesh builds a low-poly tree: a brown trunk box and two stacked green
// cones.
func treeMesh() ([]treeVertex, []uint32) {
	trunk := [3]float32{0.45, 0.32, 0.22}
	leaf := [3]float32{0.18, 0.42, 0.20}
	var verts []treeVertex
	var idx []uint32

	addTri := func(a, b, c [3]float32, col [3]float32) {
		ux := [3]float32{b[0] - a[0], b[1] - a[1], b[2] - a[2]}
		vx := [3]float32{c[0] - a[0], c[1] - a[1], c[2] - a[2]}
		n := [3]float32{ux[1]*vx[2] - ux[2]*vx[1], ux[2]*vx[0] - ux[0]*vx[2], ux[0]*vx[1] - ux[1]*vx[0]}
		l := float32(math.Sqrt(float64(n[0]*n[0] + n[1]*n[1] + n[2]*n[2])))
		if l > 1e-6 {
			n[0], n[1], n[2] = n[0]/l, n[1]/l, n[2]/l
		}
		base := uint32(len(verts))
		for _, p := range [][3]float32{a, b, c} {
			verts = append(verts, treeVertex{px: p[0], py: p[1], pz: p[2], nx: n[0], ny: n[1], nz: n[2], cr: col[0], cg: col[1], cb: col[2]})
		}
		idx = append(idx, base, base+1, base+2)
	}

	// Trunk as a thin box from y=0 to y=0.6.
	tw := float32(0.12)
	trunkTop := float32(0.6)
	box := [][3]float32{{-tw, 0, -tw}, {tw, 0, -tw}, {tw, 0, tw}, {-tw, 0, tw}}
	for i := 0; i < 4; i++ {
		a := box[i]
		b := box[(i+1)%4]
		at := [3]float32{a[0], trunkTop, a[2]}
		bt := [3]float32{b[0], trunkTop, b[2]}
		addTri(a, b, bt, trunk)
		addTri(a, bt, at, trunk)
	}

	// Two cones of foliage.
	cone := func(baseY, topY, radius float32) {
		const seg = 10
		apex := [3]float32{0, topY, 0}
		for i := 0; i < seg; i++ {
			a0 := float64(i) / seg * 2 * math.Pi
			a1 := float64(i+1) / seg * 2 * math.Pi
			p0 := [3]float32{float32(math.Cos(a0)) * radius, baseY, float32(math.Sin(a0)) * radius}
			p1 := [3]float32{float32(math.Cos(a1)) * radius, baseY, float32(math.Sin(a1)) * radius}
			addTri(p0, p1, apex, leaf)
		}
	}
	cone(0.5, 1.7, 0.6)
	cone(1.3, 2.4, 0.4)
	return verts, idx
}

// scatterTrees places trees on grassy, low-slope terrain. Positions are chosen
// from a deterministic hash so the forest is reproducible.
func scatterTrees(t *Terrain, count int) []treeInstance {
	out := make([]treeInstance, 0, count)
	half := t.WorldSize * 0.48
	hs := t.HeightScale
	for i := 0; i < count*3 && len(out) < count; i++ {
		rx := hash(i, 1)*2 - 1
		rz := hash(i, 2)*2 - 1
		x := rx * half
		z := rz * half
		h, slope := t.SampleAt(x, z)
		hN := h/hs*0.5 + 0.5
		if hN < 0.32 || hN > 0.62 || slope < 0.80 {
			continue // skip water, snow line, and steep ground
		}
		s := 1.6 + hash(i, 3)*1.4
		tint := 0.8 + hash(i, 4)*0.5
		out = append(out, treeInstance{
			ox: x, oy: h - 0.1, oz: z,
			tr: tint * 0.9, tg: tint, tb: tint * 0.8,
			scale: s,
		})
	}
	return out
}
