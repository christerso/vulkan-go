#version 450

layout(location = 0) in vec3 vNormal;
layout(location = 1) in vec3 vWorld;

layout(set = 0, binding = 0) uniform UBO {
    mat4 viewProj;
    vec4 camPos;
    vec4 lightDir;
    vec4 params;
    vec4 skyTop;
    vec4 skyHorizon;
} ubo;

layout(location = 0) out vec4 outColor;

vec3 ramp(float h, float slope) {
    vec3 water = vec3(0.10, 0.26, 0.55);
    vec3 sand  = vec3(0.76, 0.70, 0.45);
    vec3 grass = vec3(0.22, 0.46, 0.18);
    vec3 rock  = vec3(0.40, 0.36, 0.32);
    vec3 snow  = vec3(0.95, 0.96, 1.00);
    vec3 c;
    if (h < 0.30)      c = mix(water, sand, smoothstep(0.22, 0.30, h));
    else if (h < 0.48) c = mix(sand, grass, smoothstep(0.30, 0.48, h));
    else if (h < 0.70) c = mix(grass, rock, smoothstep(0.48, 0.70, h));
    else               c = mix(rock, snow, smoothstep(0.70, 0.90, h));
    // Steep faces are rock regardless of height.
    c = mix(rock, c, smoothstep(0.45, 0.75, slope));
    // Snow only settles on near-flat high ground.
    if (h > 0.62) c = mix(c, snow, smoothstep(0.80, 0.96, slope) * smoothstep(0.62, 0.78, h));
    return c;
}

void main() {
    vec3 n = normalize(vNormal);
    vec3 l = normalize(ubo.lightDir.xyz);
    vec3 v = normalize(ubo.camPos.xyz - vWorld);
    vec3 hvec = normalize(l + v);

    float hN = clamp(vWorld.y / ubo.params.x * 0.5 + 0.5, 0.0, 1.0);
    vec3 base = ramp(hN, n.y);

    float diff = max(dot(n, l), 0.0);
    float spec = pow(max(dot(n, hvec), 0.0), 24.0) * 0.15 * smoothstep(0.6, 0.95, hN);
    vec3 sky = mix(ubo.skyHorizon.rgb, ubo.skyTop.rgb, 0.5);
    vec3 ambient = base * (0.25 + 0.15 * sky);
    vec3 col = ambient + base * diff * 0.9 + spec;

    // Distance fog toward the horizon color.
    float dist = length(ubo.camPos.xyz - vWorld);
    float fog = 1.0 - exp(-dist * ubo.params.w);
    col = mix(col, ubo.skyHorizon.rgb, fog);

    outColor = vec4(col, 1.0);
}
