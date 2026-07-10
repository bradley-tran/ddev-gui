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
  const html: string[] = []
  let openCount = 0
  let index = 0
  const len = str.length

  while (index < len) {
    let nextIndex = index
    while (nextIndex < len) {
      const code = str.charCodeAt(nextIndex)
      if (code === 27 || code === 60 || code === 62 || code === 38) break
      nextIndex++
    }

    if (nextIndex > index) {
      html.push(str.slice(index, nextIndex))
      index = nextIndex
    }

    if (index >= len) break

    if (str.charCodeAt(index) === 27 && str.charCodeAt(index + 1) === 91) {
      let cursor = index + 2
      while (cursor < len && str.charCodeAt(cursor) !== 109) cursor++
      if (cursor < len) {
        while (openCount > 0) {
          html.push('</span>')
          openCount--
        }

        let codeStart = index + 2
        for (let i = codeStart; i <= cursor; i++) {
          if (i === cursor || str.charCodeAt(i) === 59) {
            if (i > codeStart) {
              const code = str.slice(codeStart, i)
              const color = COLORS[code]
              if (color) {
                html.push(`<span style="color:${color}">`)
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

    const code = str.charCodeAt(index)
    if (code === 60) html.push('&lt;')
    else if (code === 62) html.push('&gt;')
    else if (code === 38) html.push('&amp;')
    else {
      const char = str[index]
      if (char) html.push(char)
    }

    index++
  }

  while (openCount > 0) {
    html.push('</span>')
    openCount--
  }
  return html.join('')
}

export function escapeHtml(value: string): string {
  return String(value).replace(/[&<>]/g, (match) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' })[match] ?? match)
}