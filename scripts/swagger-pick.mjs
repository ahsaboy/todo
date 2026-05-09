#!/usr/bin/env node

import fs from 'node:fs'
import path from 'node:path'

const args = parseArgs(process.argv.slice(2))
const swaggerFile = path.resolve(process.cwd(), args.file || 'docs/swagger.json')
const swagger = JSON.parse(fs.readFileSync(swaggerFile, 'utf8'))

if (args.list) {
  const endpoints = []
  for (const [apiPath, pathItem] of Object.entries(swagger.paths || {})) {
    for (const [method, operation] of Object.entries(pathItem || {})) {
      endpoints.push({
        method: method.toUpperCase(),
        path: apiPath,
        summary: operation.summary || '',
        tags: operation.tags || [],
      })
    }
  }
  print(endpoints)
  process.exit(0)
}

if (args.def) {
  const def = swagger.definitions?.[args.def]
  if (!def) {
    exitWithError(`Definition not found: ${args.def}`)
  }
  print(resolveSchema(def))
  process.exit(0)
}

if (!args.path || !args.method) {
  exitWithError('Usage: node scripts/swagger-pick.mjs --path /api/v1/tasks --method get --fields summary,parameters,responses,definitions')
}

const operation = swagger.paths?.[args.path]?.[args.method.toLowerCase()]
if (!operation) {
  exitWithError(`Operation not found: ${args.method.toUpperCase()} ${args.path}`)
}

const fields = new Set((args.fields || 'summary,tags,parameters,responses,definitions').split(',').map((field) => field.trim()).filter(Boolean))
const refs = new Set()
const picked = {}

if (fields.has('summary')) {
  picked.summary = operation.summary || ''
  picked.description = operation.description || ''
}

if (fields.has('tags')) {
  picked.tags = operation.tags || []
}

if (fields.has('parameters')) {
  picked.parameters = (operation.parameters || []).map((parameter) => {
    collectRefs(parameter, refs)
    return normalizeParameter(parameter)
  })
}

if (fields.has('responses')) {
  picked.responses = {}
  for (const [status, response] of Object.entries(operation.responses || {})) {
    collectRefs(response, refs)
    picked.responses[status] = {
      description: response.description || '',
      schema: response.schema ? resolveSchema(response.schema, refs) : undefined,
    }
  }
}

if (fields.has('definitions')) {
  picked.definitions = {}
  for (const ref of refs) {
    const name = ref.replace('#/definitions/', '')
    if (swagger.definitions?.[name]) {
      picked.definitions[name] = resolveSchema(swagger.definitions[name], refs)
    }
  }
}

print({
  method: args.method.toUpperCase(),
  path: args.path,
  ...picked,
})

function parseArgs(argv) {
  const result = {}
  for (let index = 0; index < argv.length; index += 1) {
    const token = argv[index]
    if (!token.startsWith('--')) {
      continue
    }
    const key = token.slice(2)
    const next = argv[index + 1]
    if (!next || next.startsWith('--')) {
      result[key] = true
    } else {
      result[key] = next
      index += 1
    }
  }
  return result
}

function normalizeParameter(parameter) {
  return {
    name: parameter.name,
    in: parameter.in,
    required: Boolean(parameter.required),
    type: parameter.type,
    description: parameter.description || '',
    schema: parameter.schema ? resolveSchema(parameter.schema) : undefined,
  }
}

function resolveSchema(schema, refs = new Set(), seen = new Set()) {
  if (!schema || typeof schema !== 'object') {
    return schema
  }

  if (schema.$ref) {
    refs.add(schema.$ref)
    const name = schema.$ref.replace('#/definitions/', '')
    if (seen.has(name)) {
      return { $ref: schema.$ref }
    }
    const target = swagger.definitions?.[name]
    if (!target) {
      return { $ref: schema.$ref }
    }
    return {
      $ref: schema.$ref,
      resolved: resolveSchema(target, refs, new Set([...seen, name])),
    }
  }

  if (schema.allOf) {
    return {
      allOf: schema.allOf.map((item) => resolveSchema(item, refs, seen)),
    }
  }

  if (schema.type === 'array') {
    return {
      type: 'array',
      items: resolveSchema(schema.items, refs, seen),
    }
  }

  const normalized = {}
  for (const [key, value] of Object.entries(schema)) {
    if (key === 'properties') {
      normalized.properties = {}
      for (const [propName, propSchema] of Object.entries(value || {})) {
        normalized.properties[propName] = resolveSchema(propSchema, refs, seen)
      }
    } else if (key === 'items' || key === 'schema') {
      normalized[key] = resolveSchema(value, refs, seen)
    } else {
      normalized[key] = value
    }
  }
  return normalized
}

function collectRefs(value, refs) {
  if (!value || typeof value !== 'object') {
    return
  }
  if (value.$ref) {
    refs.add(value.$ref)
  }
  for (const child of Object.values(value)) {
    if (Array.isArray(child)) {
      child.forEach((item) => collectRefs(item, refs))
    } else if (child && typeof child === 'object') {
      collectRefs(child, refs)
    }
  }
}

function print(value) {
  console.log(JSON.stringify(value, null, 2))
}

function exitWithError(message) {
  console.error(message)
  process.exit(1)
}
