#version 450

layout(location = 0) in float vNdcY;

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
    // Vulkan NDC y = -1 is the top of the screen.
    float up = clamp(-vNdcY * 0.5 + 0.5, 0.0, 1.0);
    vec3 col = mix(ubo.skyHorizon.rgb, ubo.skyTop.rgb, pow(up, 0.8));
    outColor = vec4(col, 1.0);
}
