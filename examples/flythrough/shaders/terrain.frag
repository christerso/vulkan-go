#version 450

layout(location = 0) in vec3 vNormal;
layout(location = 1) in float vHeightN;

layout(set = 0, binding = 0) uniform UBO {
    mat4 viewProj;
    vec4 lightDir;
    vec4 params;
} ubo;

layout(location = 0) out vec4 outColor;

vec3 ramp(float h) {
    vec3 water = vec3(0.10, 0.26, 0.55);
    vec3 sand  = vec3(0.76, 0.70, 0.45);
    vec3 grass = vec3(0.20, 0.45, 0.18);
    vec3 rock  = vec3(0.42, 0.37, 0.33);
    vec3 snow  = vec3(0.92, 0.94, 0.98);
    if (h < 0.30) return mix(water, sand, smoothstep(0.22, 0.30, h));
    if (h < 0.50) return mix(sand, grass, smoothstep(0.30, 0.50, h));
    if (h < 0.74) return mix(grass, rock, smoothstep(0.50, 0.74, h));
    return mix(rock, snow, smoothstep(0.74, 0.94, h));
}

void main() {
    vec3 n = normalize(vNormal);
    vec3 l = normalize(ubo.lightDir.xyz);
    float diff = max(dot(n, l), 0.0) * 0.85 + 0.15;
    outColor = vec4(ramp(vHeightN) * diff, 1.0);
}
