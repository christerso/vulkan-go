#version 450

layout(location = 0) in vec3 vNormal;
layout(location = 1) in vec3 vColor;
layout(location = 2) in vec3 vWorld;

layout(set = 0, binding = 0) uniform UBO {
    mat4 viewProj;
    vec4 camPos;
    vec4 lightDir;
    vec4 params;
    vec4 skyTop;
    vec4 skyHorizon;
} ubo;

layout(location = 0) out vec4 outColor;

void main() {
    vec3 n = normalize(vNormal);
    vec3 l = normalize(ubo.lightDir.xyz);
    float diff = max(dot(n, l), 0.0) * 0.85 + 0.15;
    vec3 col = vColor * diff;

    float dist = length(ubo.camPos.xyz - vWorld);
    float fog = 1.0 - exp(-dist * ubo.params.w);
    col = mix(col, ubo.skyHorizon.rgb, fog);

    outColor = vec4(col, 1.0);
}
