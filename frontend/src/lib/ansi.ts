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