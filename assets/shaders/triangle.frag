#version 450

// Inputs from vertex shader
layout(location = 0) in vec3 fragColor;
layout(location = 1) in vec2 fragPos;
layout(location = 2) in float fragTime;

// Output color
layout(location = 0) out vec4 outColor;

// Uniform buffer
layout(binding = 0) uniform UniformBufferObject {
    float time;
    vec2 resolution;
    vec2 padding;
} ubo;

// Noise function for fancy effects
float random(vec2 st) {
    return fract(sin(dot(st.xy, vec2(12.9898,78.233))) * 43758.5453123);
}

// Smooth noise
float noise(vec2 st) {
    vec2 i = floor(st);
    vec2 f = fract(st);
    
    float a = random(i);
    float b = random(i + vec2(1.0, 0.0));
    float c = random(i + vec2(0.0, 1.0));
    float d = random(i + vec2(1.0, 1.0));
    
    vec2 u = f * f * (3.0 - 2.0 * f);
    
    return mix(a, b, u.x) + (c - a) * u.y * (1.0 - u.x) + (d - b) * u.x * u.y;
}

// Fractal noise
float fbm(vec2 st) {
    float value = 0.0;
    float amplitude = 0.5;
    float frequency = 0.0;
    
    for (int i = 0; i < 4; i++) {
        value += amplitude * noise(st);
        st *= 2.0;
        amplitude *= 0.5;
    }
    return value;
}

// HSV to RGB conversion
vec3 hsv2rgb(vec3 c) {
    vec4 K = vec4(1.0, 2.0 / 3.0, 1.0 / 3.0, 3.0);
    vec3 p = abs(fract(c.xxx + K.xyz) * 6.0 - K.www);
    return c.z * mix(K.xxx, clamp(p - K.xxx, 0.0, 1.0), c.y);
}

void main() {
    // Base color from vertex
    vec3 color = fragColor;
    
    // Add time-based color cycling
    float hueShift = fragTime * 0.3;
    
    // Calculate distance from center for radial effects
    float dist = length(fragPos);
    
    // Add rainbow effect based on position and time
    float hue = atan(fragPos.y, fragPos.x) / (2.0 * 3.14159) + 0.5;
    hue += fragTime * 0.1; // Rotate colors over time
    
    // Add noise-based color variation
    vec2 noiseCoord = fragPos * 3.0 + fragTime * 0.5;
    float noiseValue = fbm(noiseCoord);
    
    // Create psychedelic color effect
    vec3 rainbowColor = hsv2rgb(vec3(hue + noiseValue * 0.3, 0.8, 1.0));
    
    // Mix original color with effects
    color = mix(color, rainbowColor, 0.6);
    
    // Add shimmer effect
    float shimmer = sin(fragTime * 4.0 + dist * 10.0) * 0.5 + 0.5;
    color += vec3(shimmer * 0.2);
    
    // Add pulsing brightness
    float pulse = sin(fragTime * 2.0) * 0.1 + 1.0;
    color *= pulse;
    
    // Add edge glow effect
    float edgeGlow = 1.0 - smoothstep(0.0, 0.3, dist);
    color += vec3(edgeGlow * 0.3);
    
    // Add particles effect using noise
    vec2 particleCoord = fragPos * 8.0 + fragTime * vec2(1.0, 0.5);
    float particles = noise(particleCoord);
    particles = smoothstep(0.7, 0.8, particles);
    color += vec3(particles * 0.5);
    
    // Vignette effect
    vec2 screenPos = gl_FragCoord.xy / ubo.resolution.xy;
    float vignette = distance(screenPos, vec2(0.5));
    vignette = 1.0 - vignette;
    vignette = smoothstep(0.0, 1.0, vignette);
    color *= vignette;
    
    // Final color output with alpha
    float alpha = 1.0;
    
    // Add transparency effect at edges
    alpha *= smoothstep(0.8, 0.0, dist);
    
    outColor = vec4(color, alpha);
}