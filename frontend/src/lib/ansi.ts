const COLORS: Record<string, string> = {
  '30': '#4e4e4e',
  '31': '#e06c75',
  '32': '#98c379',
  '33': '#e5c07b',
  '34': '#61afef',
  '35': '#c678dd',
  '36': '#56b6c2',
  '37': '#dcdfe4',
  '90': '#7f8490',
  '91': '#e06c75',
  '92': '#98c379',
  '93': '#e5c07b',
  '94': '#61afef',
  '95': '#c678dd',
  '96': '#56b6c2',
  '97': '#ffffff',
}

function splitTrailingPunctuation(rawUrl: string): [string, string] {
  let end = rawUrl.length

  while (end > 0) {
    const lastChar = rawUrl[end - 1]
    const isTrailingPunctuation = lastChar === ')' || lastChar === ',' || lastChar === '.' || lastChar === ';' || lastChar === '!' || lastChar === '?'
    if (!isTrailingPunctuation) break

    if (lastChar === ')') {
      const candidate = rawUrl.slice(0, end)
      const opens = (candidate.match(/\(/g) ?? []).length
      const closes = (candidate.match(/\)/g) ?? []).length
      if (closes <= opens) break
    }

    end--
  }

  return [rawUrl.slice(0, end), rawUrl.slice(end)]
}

function linkifyTextSegment(segment: string): string {
  return segment.replace(/\bhttps?:\/\/[^\s<>"']+/gi, (rawUrl) => {
    const [url, trailing] = splitTrailingPunctuation(rawUrl)
    if (!url) return rawUrl

    return `<a href="${url}" class="terminal-link" data-terminal-url="${url}" target="_blank" rel="noopener noreferrer">${url}</a>${trailing}`
  })
}

export function linkifyHtmlUrls(html: string): string {
  if (!html) return ''

  return html
    .split(/(<[^>]+>)/g)
    .map((segment) => {
      if (segment.startsWith('<') && segment.endsWith('>')) return segment
      return linkifyTextSegment(segment)
    })
    .join('')
}

export function ansiToHtml(str: string): string {
  let html = ''
  let openCount = 0
  let index = 0

  while (index < str.length) {
    if (str[index] === '\x1b' && str[index + 1] === '[') {
      let cursor = index + 2
      while (cursor < str.length && str[cursor] !== 'm') cursor++
      if (cursor < str.length) {
        while (openCount > 0) {
          html += '</span>'
          openCount--
        }

        let codeStart = index + 2
        for (let i = codeStart; i <= cursor; i++) {
          if (i === cursor || str[i] === ';') {
            if (i > codeStart) {
              const code = str.slice(codeStart, i)
              const color = COLORS[code]
              if (color) {
                html += `<span style="color:${color}">`
                openCount++
              }
            }
            codeStart = i + 1
          }
        }

        index = cursor + 1
        continue
      }
    }

    const char = str[index]
    if (char === '<') html += '&lt;'
    else if (char === '>') html += '&gt;'
    else if (char === '&') html += '&amp;'
    else html += char

    index++
  }

  while (openCount > 0) {
    html += '</span>'
    openCount--
  }
  return html
}

export function escapeHtml(value: string): string {
  return String(value).replace(/[&<>]/g, (match) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' })[match] ?? match)
}