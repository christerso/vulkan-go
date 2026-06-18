#version 450

layout(location = 0) in vec3 inPos;
layout(location = 1) in vec3 inNormal;

layout(set = 0, binding = 0) uniform UBO {
    mat4 viewProj;
    vec4 camPos;     // xyz camera position
    vec4 lightDir;   // xyz direction toward the sun
    vec4 params;     // x=heightScale y=time z=seaLevel w=fogDensity
    vec4 skyTop;
    vec4 skyHorizon;
} ubo;

layout(location = 0) out vec3 vNormal;
layout(location = 1) out vec3 vWorld;

void main() {
    gl_Position = ubo.viewProj * vec4(inPos, 1.0);
    vNormal = inNormal;
    vWorld = inPos;
}
