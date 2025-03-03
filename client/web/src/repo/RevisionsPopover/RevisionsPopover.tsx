import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useEffect } from 'react'

import { GitRefType, Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { Button, useLocalStorage, Tab, TabList, TabPanel, TabPanels } from '@sourcegraph/wildcard'

import { GitCommitAncestorFields, GitRefFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { replaceRevisionInURL } from '../../util/url'

import { ConnectionPopoverTabs } from './components'
import styles from './RevisionsPopover.module.scss'
import { RevisionsPopoverCommits } from './RevisionsPopoverCommits'
import { RevisionsPopoverReferences } from './RevisionsPopoverReferences'

export interface RevisionsPopoverProps {
    repo: Scalars['ID']
    repoName: string
    defaultBranch: string

    /** The current revision, or undefined for the default branch. */
    currentRev: string | undefined

    currentCommitID?: string

    /* Callback to dismiss the parent popover wrapper */
    togglePopover: () => void

    /* Determine the URL to use for each revision node */
    getPathFromRevision?: (href: string, revision: string) => string

    /**
     * If the popover should display result nodes that are not **known** revisions.
     * This ensures we can support ancestory-based revision queries (e.g. `main^1`).
     */
    showSpeculativeResults?: boolean

    /**
     * The selected revision node. Should be used to trigger side effects from clicking a node, e.g. calling `eventLogger`.
     */
    onSelect?: (node: GitRefFields | GitCommitAncestorFields) => void
}

type RevisionsPopoverTabID = 'branches' | 'tags' | 'commits'

interface RevisionsPopoverTab {
    id: RevisionsPopoverTabID
    label: string
    noun: string
    pluralNoun: string
    type?: GitRefType
}

const LAST_TAB_STORAGE_KEY = 'RevisionsPopover.lastTab'

const TABS: RevisionsPopoverTab[] = [
    { id: 'branches', label: 'Branches', noun: 'branch', pluralNoun: 'branches', type: GitRefType.GIT_BRANCH },
    { id: 'tags', label: 'Tags', noun: 'tag', pluralNoun: 'tags', type: GitRefType.GIT_TAG },
    { id: 'commits', label: 'Commits', noun: 'commit', pluralNoun: 'commits' },
]

/**
 * A popover that displays a searchable list of revisions (grouped by type) for
 * the current repository.
 */
export const RevisionsPopover: React.FunctionComponent<RevisionsPopoverProps> = props => {
    const { getPathFromRevision = replaceRevisionInURL } = props

    useEffect(() => {
        eventLogger.logViewEvent('RevisionsPopover')
    }, [])

    const [tabIndex, setTabIndex] = useLocalStorage(LAST_TAB_STORAGE_KEY, 0)
    const handleTabsChange = useCallback((index: number) => setTabIndex(index), [setTabIndex])

    return (
        <ConnectionPopoverTabs
            className={styles.revisionsPopover}
            data-testid="revisions-popover"
            defaultIndex={tabIndex}
            onChange={handleTabsChange}
        >
            <TabList
                wrapperClassName={styles.tabs}
                actions={
                    <Button
                        onClick={props.togglePopover}
                        variant="icon"
                        className={styles.tabsClose}
                        aria-label="Close"
                    >
                        <CloseIcon className="icon-inline" />
                    </Button>
                }
            >
                {TABS.map(({ label, id }) => (
                    <Tab key={id} data-tab-content={id}>
                        <span className="tablist-wrapper--tab-label">{label}</span>
                    </Tab>
                ))}
            </TabList>
            <TabPanels>
                {TABS.map(tab => (
                    <TabPanel key={tab.id}>
                        {tab.type ? (
                            <RevisionsPopoverReferences
                                noun={tab.noun}
                                pluralNoun={tab.pluralNoun}
                                type={tab.type}
                                currentRev={props.currentRev}
                                getPathFromRevision={getPathFromRevision}
                                defaultBranch={props.defaultBranch}
                                repo={props.repo}
                                repoName={props.repoName}
                                onSelect={props.onSelect}
                                showSpeculativeResults={
                                    props.showSpeculativeResults && tab.type === GitRefType.GIT_BRANCH
                                }
                            />
                        ) : (
                            <RevisionsPopoverCommits
                                noun={tab.noun}
                                pluralNoun={tab.pluralNoun}
                                currentRev={props.currentRev}
                                getPathFromRevision={getPathFromRevision}
                                defaultBranch={props.defaultBranch}
                                repo={props.repo}
                                currentCommitID={props.currentCommitID}
                                onSelect={props.onSelect}
                            />
                        )}
                    </TabPanel>
                ))}
            </TabPanels>
        </ConnectionPopoverTabs>
    )
}
