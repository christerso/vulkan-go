#version 450

// Vertex attributes
layout(location = 0) in vec2 inPosition;
layout(location = 1) in vec3 inColor;

// Uniform buffer with animation data
layout(binding = 0) uniform UniformBufferObject {
    float time;
    vec2 resolution;
    vec2 padding;
} ubo;

// Outputs to fragment shader
layout(location = 0) out vec3 fragColor;
layout(location = 1) out vec2 fragPos;
layout(location = 2) out float fragTime;

void main() {
    // Pass time to fragment shader for animation
    fragTime = ubo.time;
    
    // Calculate animated position
    vec2 pos = inPosition;
    
    // Add subtle breathing animation
    float breathScale = 1.0 + 0.05 * sin(ubo.time * 2.0);
    pos *= breathScale;
    
    // Add rotation animation
    float angle = ubo.time * 0.5;
    float cosAngle = cos(angle);
    float sinAngle = sin(angle);
    mat2 rotation = mat2(cosAngle, -sinAngle, sinAngle, cosAngle);
    pos = rotation * pos;
    
    // Add subtle floating motion
    pos.y += 0.1 * sin(ubo.time * 1.5);
    pos.x += 0.05 * cos(ubo.time * 1.2);
    
    gl_Position = vec4(pos, 0.0, 1.0);
    
    // Pass animated color to fragment shader
    fragColor = inColor;
    
    // Pass position for effects
    fragPos = pos;
}