#version 450

layout(location = 0) in vec3 inPos;
layout(location = 1) in vec3 inNormal;
layout(location = 2) in vec3 inColor;
layout(location = 3) in vec3 iOffset;
layout(location = 4) in vec3 iTint;
layout(location = 5) in float iScale;

layout(set = 0, binding = 0) uniform UBO {
    mat4 viewProj;
    vec4 camPos;
    vec4 lightDir;
    vec4 params;
    vec4 skyTop;
    vec4 skyHorizon;
} ubo;

layout(location = 0) out vec3 vNormal;
layout(location = 1) out vec3 vColor;
layout(location = 2) out vec3 vWorld;

void main() {
    vec3 world = inPos * iScale + iOffset;
    gl_Position = ubo.viewProj * vec4(world, 1.0);
    vNormal = inNormal;
    vColor = inColor * iTint;
    vWorld = world;
}
