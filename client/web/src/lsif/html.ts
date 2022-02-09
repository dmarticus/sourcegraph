import { JsonDocument, Occurrence, SyntaxKind } from './spec'

class HtmlBuilder {
    public readonly buffer: string[] = []
    public plaintext(value: string): void {
        if (!value) {
            return
        }

        this.span('', value)
    }
    public span(attributes: string, value: string): void {
        this.element('span', attributes, value)
    }
    public element(element: string, attributes: string, value: string): void {
        this.openTag(element + ' ' + attributes)
        this.raw(value)
        this.closeTag(element)
    }
    public raw(html: string): void {
        this.buffer.push(html)
    }
    public openTag(tag: string): void {
        this.buffer.push('<')
        this.buffer.push(tag)
        this.buffer.push('>')
    }
    public closeTag(tag: string): void {
        this.buffer.push('</')
        this.buffer.push(tag)
        this.buffer.push('>')
    }
}

function openLine(html: HtmlBuilder, language: string, lineNumber: number) {
    html.openTag('tr')
    html.raw(`<td class="line" data-line="${lineNumber + 1}"></td>`)

    html.openTag('td class="code"')
    html.openTag('div')
    html.openTag(`span class="hl-source hl-${language}"`)
}

function closeLine(html: HtmlBuilder) {
    html.closeTag('span')
    html.closeTag('div')
    html.closeTag('td')
    html.closeTag('tr')
}

function highlightSlice(html: HtmlBuilder, kind: SyntaxKind, slice: string) {
    let kindName = SyntaxKind[kind]
    if (kindName) {
        html.span(`class="hl-typed-${kindName}"`, slice)
    } else {
        html.plaintext(slice)
    }
}

// Currently assumes that no ranges overlap in the occurences.
export function render(lsif_json: string, content: string): string {
    const language = 'go'

    let occurrences = (JSON.parse(lsif_json) as JsonDocument).occurrences.map(occ => new Occurrence(occ))

    // Sort by line, and then by start character.
    occurrences.sort((a, b) => {
        if (a.range.start.line != b.range.start.line) {
            return a.range.start.line - b.range.start.line
        }

        return a.range.start.character - b.range.start.character
    })

    const lines = content.replaceAll('\r\n', '\n').split('\n')
    const html = new HtmlBuilder()

    let occIdx = 0

    html.openTag('table')
    html.openTag('tbody')
    for (let lineNumber = 0; lineNumber < lines.length; lineNumber++) {
        openLine(html, language, lineNumber)

        let line = lines[lineNumber]

        let startCharacter = 0
        while (occIdx < occurrences.length && occurrences[occIdx].range.start.line == lineNumber) {
            let occ = occurrences[occIdx]
            occIdx++

            let { start, end } = occ.range

            html.plaintext(line.slice(startCharacter, start.character))
            if (start.line != end.line) {
                html.plaintext(line.slice(start.character))
                closeLine(html)

                // Move to the next line
                lineNumber++

                // Handle all the lines that completely owned by this occurence
                while (lineNumber < end.line) {
                    line = lines[lineNumber]

                    openLine(html, language, lineNumber)
                    highlightSlice(html, occ.kind, lines[lineNumber])
                    closeLine(html)

                    lineNumber++
                }

                // Write as much of the line as the last occurence owns
                line = lines[lineNumber]

                openLine(html, language, lineNumber)
                highlightSlice(html, occ.kind, line.slice(0, end.character))
            } else {
                highlightSlice(html, occ.kind, line.slice(start.character, end.character))
            }

            startCharacter = end.character
        }

        // If we didn't find any occurences on this line, then just write the line plainly
        if (startCharacter == 0) {
            html.plaintext(line)
        }
        closeLine(html)
    }
    html.closeTag('tbody')
    html.closeTag('table')

    return html.buffer.join('')
}
