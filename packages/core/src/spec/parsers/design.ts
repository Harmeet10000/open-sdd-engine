export interface MermaidBlock {
  startLine: number;
  valid: boolean;
}

export function parseDesign(content: string): {
  hasOverview: boolean;
  hasArchitecture: boolean;
  hasFileStructure: boolean;
  mermaidBlocks: MermaidBlock[];
} {
  const hasOverview = /^##\s+Overview/m.test(content);
  const hasArchitecture = /^##\s+Architecture/m.test(content);
  const hasFileStructure = /^##\s+(File Structure|Files)/im.test(content);

  const mermaidBlocks: MermaidBlock[] = [];
  const lines = content.split("\n");
  let inMermaid = false;
  let startLine = 0;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]!;
    if (line.trim() === "```mermaid") {
      inMermaid = true;
      startLine = i + 1;
    } else if (inMermaid && line.trim() === "```") {
      mermaidBlocks.push({ startLine, valid: i > startLine });
      inMermaid = false;
    }
  }
  if (inMermaid) {
    mermaidBlocks.push({ startLine, valid: false });
  }

  return { hasOverview, hasArchitecture, hasFileStructure, mermaidBlocks };
}
