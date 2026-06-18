#version 450

layout(location = 0) in vec3 inPos;
layout(location = 1) in vec3 inNormal;

layout(set = 0, binding = 0) uniform UBO {
    mat4 viewProj;
    vec4 lightDir;   // xyz direction
    vec4 params;     // x = heightScale
} ubo;

layout(location = 0) out vec3 vNormal;
layout(location = 1) out float vHeightN;

void main() {
    gl_Position = ubo.viewProj * vec4(inPos, 1.0);
    vNormal = inNormal;
    vHeightN = clamp(inPos.y / ubo.params.x * 0.5 + 0.5, 0.0, 1.0);
}
