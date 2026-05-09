#!/usr/bin/env node

import fs from 'node:fs'
import path from 'node:path'

const root = path.resolve(process.cwd(), 'src')
const forbiddenPatterns = [
  {
    pattern: /(['"`])\/api\/v1(?:\/|\1)/g,
    message: "Do not use '/api/v1'. Use the relative API base 'api/v1'.",
  },
  {
    pattern: /location\.origin\s*\+\s*(['"`])\/api\/v1/g,
    message: "Do not build API URLs from location.origin + '/api/v1'. Use 'api/v1'.",
  },
]

const files = collectFiles(root, new Set(['.ts', '.tsx', '.vue', '.js', '.jsx']))
const violations = []

for (const file of files) {
  const content = fs.readFileSync(file, 'utf8')
  for (const { pattern, message } of forbiddenPatterns) {
    pattern.lastIndex = 0
    let match
    while ((match = pattern.exec(content)) !== null) {
      const { line, column } = getLocation(content, match.index)
      violations.push({
        file: path.relative(process.cwd(), file),
        line,
        column,
        message,
      })
    }
  }
}

if (violations.length > 0) {
  console.error('Forbidden API path usage found:')
  for (const violation of violations) {
    console.error(`- ${violation.file}:${violation.line}:${violation.column} ${violation.message}`)
  }
  process.exit(1)
}

console.log("API path check passed. Use 'api/v1' for relative API requests.")

function collectFiles(dir, extensions) {
  if (!fs.existsSync(dir)) {
    return []
  }

  const result = []
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const fullPath = path.join(dir, entry.name)
    if (entry.isDirectory()) {
      result.push(...collectFiles(fullPath, extensions))
      continue
    }
    if (extensions.has(path.extname(entry.name))) {
      result.push(fullPath)
    }
  }
  return result
}

function getLocation(content, index) {
  const before = content.slice(0, index)
  const lines = before.split(/\r?\n/)
  return {
    line: lines.length,
    column: lines.at(-1).length + 1,
  }
}
