const IMAGE_EXTS = new Set(['png', 'jpg', 'jpeg', 'gif', 'svg', 'webp', 'ico', 'bmp', 'avif'])

const BINARY_EXTS = new Set([
  'zip', 'tar', 'gz', 'bz2', 'xz', '7z', 'rar', 'tgz',
  'exe', 'dll', 'so', 'dylib', 'o', 'a', 'class', 'pyc', 'phar', 'wasm',
  'mp3', 'mp4', 'wav', 'ogg', 'flac', 'avi', 'mov', 'mkv', 'webm',
  'woff', 'woff2', 'ttf', 'otf', 'eot',
  'pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx',
  'db', 'sqlite', 'sqlite3',
  'bin', 'dat', 'iso', 'img', 'dmg',
])

const EXT_LANG: Record<string, string> = {
  js: 'javascript',
  jsx: 'javascript',
  ts: 'typescript',
  tsx: 'typescript',
  py: 'python',
  rb: 'ruby',
  go: 'go',
  rs: 'rust',
  java: 'java',
  kt: 'kotlin',
  cs: 'csharp',
  c: 'c',
  cpp: 'cpp',
  h: 'c',
  hpp: 'cpp',
  php: 'php',
  swift: 'swift',
  sh: 'bash',
  bash: 'bash',
  zsh: 'bash',
  ps1: 'powershell',
  sql: 'sql',
  html: 'xml',
  htm: 'xml',
  xml: 'xml',
  svg: 'xml',
  css: 'css',
  scss: 'scss',
  less: 'less',
  json: 'json',
  yaml: 'yaml',
  yml: 'yaml',
  toml: 'ini',
  ini: 'ini',
  env: 'bash',
  dockerfile: 'dockerfile',
  makefile: 'makefile',
  lua: 'lua',
  r: 'r',
  dart: 'dart',
  vue: 'xml',
  graphql: 'graphql',
  gql: 'graphql',
  tf: 'hcl',
  hcl: 'hcl',
  proto: 'protobuf',
  twig: 'twig',
}

const NAME_LANG: Record<string, string> = {
  dockerfile: 'dockerfile',
  makefile: 'makefile',
  vagrantfile: 'ruby',
  gemfile: 'ruby',
  rakefile: 'ruby',
  '.gitignore': 'bash',
  '.dockerignore': 'bash',
  '.editorconfig': 'ini',
  '.htaccess': 'apache',
  '.env': 'bash',
  '.env.example': 'bash',
}

const MIME_BY_EXT: Record<string, string> = {
  png: 'image/png',
  jpg: 'image/jpeg',
  jpeg: 'image/jpeg',
  gif: 'image/gif',
  svg: 'image/svg+xml',
  webp: 'image/webp',
  ico: 'image/x-icon',
  bmp: 'image/bmp',
  avif: 'image/avif',
}

const CODE_EXTS = new Set(Object.keys(EXT_LANG))

function extensionOf(name: string): string {
  if (!name.includes('.')) return ''
  return name.toLowerCase().split('.').pop() || ''
}

export function isMarkdownFile(name: string): boolean {
  const lower = name.toLowerCase()
  return lower.endsWith('.md') || lower.endsWith('.markdown') || lower.endsWith('.mdx')
}

export function isImageFile(name: string): boolean {
  return IMAGE_EXTS.has(extensionOf(name))
}

export function isBinaryFile(name: string): boolean {
  return BINARY_EXTS.has(extensionOf(name))
}

export function isCodeFile(name: string): boolean {
  const lowerName = name.toLowerCase()
  return CODE_EXTS.has(extensionOf(name)) || lowerName in NAME_LANG
}

export function detectLanguage(name: string): string {
  const lowerName = name.toLowerCase()
  const mappedLanguage = NAME_LANG[lowerName]
  if (mappedLanguage) return mappedLanguage
  return EXT_LANG[extensionOf(name)] || 'plaintext'
}

export function mimeFromExt(name: string): string {
  return MIME_BY_EXT[extensionOf(name)] || 'image/png'
}