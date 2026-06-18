#version 450

layout(location = 0) out float vNdcY;

// Fullscreen triangle from the vertex index; no vertex buffer.
void main() {
    vec2 p = vec2((gl_VertexIndex << 1) & 2, gl_VertexIndex & 2);
    vec2 ndc = p * 2.0 - 1.0;
    vNdcY = ndc.y;
    gl_Position = vec4(ndc, 1.0, 1.0);
}
