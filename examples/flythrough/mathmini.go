package main

import "math"

// vec3 is a minimal 3-vector for the demo.
type vec3 struct{ x, y, z float32 }

func sub(a, b vec3) vec3 { return vec3{a.x - b.x, a.y - b.y, a.z - b.z} }

func cross(a, b vec3) vec3 {
	return vec3{a.y*b.z - a.z*b.y, a.z*b.x - a.x*b.z, a.x*b.y - a.y*b.x}
}

func normalize(v vec3) vec3 {
	l := float32(math.Sqrt(float64(v.x*v.x + v.y*v.y + v.z*v.z)))
	if l < 1e-8 {
		return vec3{}
	}
	return vec3{v.x / l, v.y / l, v.z / l}
}

func dot(a, b vec3) float32 { return a.x*b.x + a.y*b.y + a.z*b.z }

// mat4 is column-major, matching GLSL layout.
type mat4 [16]float32

func mul(a, b mat4) mat4 {
	var r mat4
	for c := 0; c < 4; c++ {
		for row := 0; row < 4; row++ {
			r[c*4+row] = a[0*4+row]*b[c*4+0] + a[1*4+row]*b[c*4+1] + a[2*4+row]*b[c*4+2] + a[3*4+row]*b[c*4+3]
		}
	}
	return r
}

// perspective builds a Vulkan projection (Z in [0,1], Y flipped).
func perspective(fovY, aspect, near, far float32) mat4 {
	f := float32(1.0 / math.Tan(float64(fovY)/2))
	var m mat4
	m[0] = f / aspect
	m[5] = -f // flip Y for Vulkan clip space
	m[10] = far / (near - far)
	m[11] = -1
	m[14] = (near * far) / (near - far)
	return m
}

// lookAt builds a right-handed view matrix.
func lookAt(eye, center, up vec3) mat4 {
	f := normalize(sub(center, eye))
	s := normalize(cross(f, up))
	u := cross(s, f)
	var m mat4
	m[0], m[4], m[8] = s.x, s.y, s.z
	m[1], m[5], m[9] = u.x, u.y, u.z
	m[2], m[6], m[10] = -f.x, -f.y, -f.z
	m[12] = -dot(s, eye)
	m[13] = -dot(u, eye)
	m[14] = dot(f, eye)
	m[15] = 1
	return m
}
