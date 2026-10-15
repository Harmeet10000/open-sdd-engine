import { access } from "node:fs/promises";
import { constants } from "node:fs";
import { join, dirname } from "node:path";
import YAML from "yaml";

export interface OkfMetadata {
  type: string;
  title: string;
  tags?: string[];
  timestamp?: string;
  [key: string]: any;
}

export interface OkfBundle {
  path: string;
  meta: OkfMetadata;
  content: string;
  links: string[];
}

export function parseOkf(content: string): { meta: OkfMetadata | null; body: string; links: string[] } {
  const normalized = content.replace(/\r\n/g, "\n");
  const match = normalized.match(/^---\n([\s\S]*?)\n---\n([\s\S]*)$/);
  
  let meta: OkfMetadata | null = null;
  let body = content;
  
  if (match) {
    try {
      meta = YAML.parse(match[1]!) as OkfMetadata;
      body = match[2]!;
    } catch {
      // Invalid YAML frontmatter
    }
  }

  // Extract relative markdown links (ignores http/https/absolute paths/hash anchors)
  const links: string[] = [];
  const linkRegex = /\[.*?\]\((?!https?:|\/)([^)#\s]+(?:#[^)\s]+)?)\)/g;
  let m;
  while ((m = linkRegex.exec(body)) !== null) {
    if (m[1]) {
      // Strip fragment identifier
      const filePart = m[1].split("#")[0];
      if (filePart && filePart.endsWith(".md") && !links.includes(filePart)) {
        links.push(filePart);
      }
    }
  }

  return { meta, body, links };
}

export async function validateOkfFile(
  filePath: string,
  content: string,
): Promise<{ valid: boolean; errors: string[]; warnings: string[]; links: string[] }> {
  const errors: string[] = [];
  const warnings: string[] = [];
  
  const { meta, links } = parseOkf(content);
  
  if (!meta) {
    errors.push("Missing or invalid YAML frontmatter (must be enclosed in ---)");
    return { valid: false, errors, warnings, links: [] };
  }

  if (!meta.type) {
    errors.push("Frontmatter missing required field: 'type'");
  }
  if (!meta.title) {
    errors.push("Frontmatter missing required field: 'title'");
  }

  return {
    valid: errors.length === 0,
    errors,
    warnings,
    links,
  };
}

export function getOkfArchitectureTemplate(): string {
  return `---
type: architecture
title: System Architecture Guide
tags: [architecture, design, overview]
timestamp: ${new Date().toISOString()}
---

# System Architecture Guide

This guide describes the core architecture of the system.

## Components

- **Core Module**: Handles data processing and ingestion.
- **API Engine**: Serves sub-millisecond query responses. See [Development Guidelines](development.md) for contribution rules.

## Data Flow

The following data flow describes how inputs are transformed into predictions:
1. Ingest events via API.
2. Queue events for background worker.
`;
}

export function getOkfDevelopmentTemplate(): string {
  return `---
type: process
title: Development Guidelines
tags: [guide, development, onboarding]
timestamp: ${new Date().toISOString()}
---

# Development Guidelines

Welcome to the project! Please follow these standards.

## Code Standards

- Write TypeScript code targeting ES2022.
- Use Bun for local development and running test suites.
- Read [System Architecture Guide](architecture.md) before designing changes.

## Testing Flow

- Run tests with \`bun test\`.
- All features must have unit tests.
`;
}
