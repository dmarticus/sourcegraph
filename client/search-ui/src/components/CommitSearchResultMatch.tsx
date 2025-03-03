import classNames from 'classnames'
import { isEqual, range } from 'lodash'
import React from 'react'
import VisibilitySensor from 'react-visibility-sensor'
import { combineLatest, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, switchMap } from 'rxjs/operators'
import sanitizeHtml from 'sanitize-html'

import { highlightNode } from '@sourcegraph/common'
import { highlightCode } from '@sourcegraph/search'
import { LastSyncedIcon } from '@sourcegraph/shared/src/components/LastSyncedIcon'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { CommitMatch } from '@sourcegraph/shared/src/search/stream'
import { LoadingSpinner, Link } from '@sourcegraph/wildcard'

import styles from './CommitSearchResultMatch.module.scss'
import searchResultStyles from './SearchResult.module.scss'

interface CommitSearchResultMatchProps extends PlatformContextProps<'requestGraphQL'> {
    item: CommitMatch
}

interface CommitSearchResultMatchState {
    HTML?: string
}

export class CommitSearchResultMatch extends React.Component<
    CommitSearchResultMatchProps,
    CommitSearchResultMatchState
> {
    public state: CommitSearchResultMatchState = {}
    private tableContainerElement: HTMLElement | null = null
    private visibilitySensorOffset = { bottom: -500 }

    private visibilityChanges = new Subject<boolean>()
    private subscriptions = new Subscription()
    private propsChanges = new Subject<CommitSearchResultMatchProps>()

    private getLanguage(): string | undefined {
        const matches = /```(\S+)\s/.exec(this.props.item.content)
        if (!matches) {
            return undefined
        }
        return matches[1]
    }

    constructor(props: CommitSearchResultMatchProps) {
        super(props)
        // Render the match body as markdown, and syntax highlight the response if it's a code block.
        // This is a lot of network requests right now, but once extensions can run on the backend we can
        // run results through the renderer and syntax highlighter without network requests.
        this.subscriptions.add(
            combineLatest([this.propsChanges, this.visibilityChanges])
                .pipe(
                    filter(([, isVisible]) => isVisible),
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    switchMap(([props]) => {
                        const codeContent = props.item.content.replace(/^```[_a-z]*\n/i, '').replace(/```$/i, '') // Remove Markdown code indicators to render code as plain text

                        const lang = this.getLanguage() || 'txt'

                        // Match the code content and any trailing newlines if any.
                        if (codeContent) {
                            return highlightCode({
                                code: codeContent,
                                fuzzyLanguage: lang,
                                disableTimeout: false,
                                platformContext: props.platformContext,
                            }).pipe(
                                // Return the rendered markdown if highlighting fails.
                                catchError(error => {
                                    console.log(error)
                                    return of(codeContent)
                                })
                            )
                        }

                        return of(codeContent)
                    }),
                    // Return the raw body if markdown rendering fails, maintaining the text structure.
                    catchError(() => of('<pre>' + sanitizeHtml(props.item.content) + '</pre>'))
                )
                .subscribe(
                    string => this.setState({ HTML: string }),
                    error => console.error(error)
                )
        )
    }

    public componentDidMount(): void {
        this.propsChanges.next(this.props)
        this.highlightNodes()
    }

    public componentDidUpdate(): void {
        this.propsChanges.next(this.props)
        this.highlightNodes()
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private highlightNodes(): void {
        if (this.tableContainerElement) {
            const visibleRows = this.tableContainerElement.querySelectorAll('table tr')
            if (visibleRows.length > 0) {
                for (const [line, character, length] of this.props.item.ranges) {
                    const code = visibleRows[line - 1]
                    if (code) {
                        highlightNode(code as HTMLElement, character, length)
                    }
                }
            }
        }
    }

    private onChangeVisibility = (isVisible: boolean): void => {
        this.visibilityChanges.next(isVisible)
    }

    private getFirstLine(): number {
        if (this.props.item.ranges.length === 0) {
            // If there are no highlights, the calculation below results in -Infinity.
            return 0
        }
        return Math.max(0, Math.min(...this.props.item.ranges.map(([line]) => line)) - 1)
    }

    private getLastLine(): number {
        if (this.props.item.ranges.length === 0) {
            // If there are no highlights, the calculation below results in Infinity,
            // so we set lastLine to 5, which is a just a heuristic for a medium-sized result.
            return 5
        }
        const lastLine = Math.max(...this.props.item.ranges.map(([line]) => line)) + 1
        return this.props.item.ranges ? Math.min(lastLine, this.props.item.ranges.length) : lastLine
    }

    public render(): JSX.Element {
        const firstLine = this.getFirstLine()
        let lastLine = this.getLastLine()
        if (firstLine === lastLine) {
            // Some edge cases yield the same first and last line, causing the visibility sensor to break, so make sure to avoid this.
            lastLine++
        }

        return (
            <VisibilitySensor
                active={true}
                onChange={this.onChangeVisibility}
                partialVisibility={true}
                offset={this.visibilitySensorOffset}
            >
                <div className={styles.commitSearchResultMatch}>
                    {this.props.item.repoLastFetched && (
                        <LastSyncedIcon
                            className={styles.lastSyncedIcon}
                            lastSyncedTime={this.props.item.repoLastFetched}
                        />
                    )}
                    {this.state.HTML !== undefined ? (
                        <Link
                            key={this.props.item.url}
                            to={this.props.item.url}
                            className={searchResultStyles.searchResultMatch}
                        >
                            <code>
                                <Markdown
                                    refFn={this.setTableContainerElement}
                                    testId="search-result-match-code-excerpt"
                                    className={classNames(styles.markdown, styles.codeExcerpt)}
                                    dangerousInnerHTML={this.state.HTML}
                                />
                            </code>
                        </Link>
                    ) : (
                        <>
                            <LoadingSpinner className={styles.loader} />
                            <table>
                                <tbody>
                                    {range(firstLine, lastLine).map(index => (
                                        <tr key={`${this.props.item.url}#${index}`}>
                                            {/* create empty space to fill viewport (as if the blob content were already fetched, otherwise we'll overfetch) */}
                                            <td className={styles.lineHidden}>
                                                <code>{index}</code>
                                            </td>
                                            <td className="code"> </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </>
                    )}
                </div>
            </VisibilitySensor>
        )
    }

    private setTableContainerElement = (reference: HTMLElement | null): void => {
        this.tableContainerElement = reference
    }
}
