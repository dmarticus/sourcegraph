import classNames from 'classnames'
import * as H from 'history'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import SyncIcon from 'mdi-react/SyncIcon'
import React, { useState, useCallback, useEffect } from 'react'

import { ErrorAlert, ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { Hoverifier } from '@sourcegraph/codeintellify'
import { asError, isErrorLike } from '@sourcegraph/common'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { HoverMerged } from '@sourcegraph/shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { ChangesetState } from '@sourcegraph/shared/src/graphql-operations'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { RepoSpec, RevisionSpec, FileSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'
import { InputTooltip } from '@sourcegraph/web/src/components/InputTooltip'
import { Button, Alert } from '@sourcegraph/wildcard'

import { DiffStatStack } from '../../../../components/diff/DiffStat'
import { ChangesetSpecType, ExternalChangesetFields } from '../../../../graphql-operations'
import {
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
    reenqueueChangeset,
} from '../backend'

import { ChangesetCheckStatusCell } from './ChangesetCheckStatusCell'
import { ChangesetFileDiff } from './ChangesetFileDiff'
import { ChangesetReviewStatusCell } from './ChangesetReviewStatusCell'
import { ChangesetStatusCell } from './ChangesetStatusCell'
import { DownloadDiffButton } from './DownloadDiffButton'
import { ExternalChangesetInfoCell } from './ExternalChangesetInfoCell'
import styles from './ExternalChangesetNode.module.scss'

export interface ExternalChangesetNodeProps extends ThemeProps {
    node: ExternalChangesetFields
    viewerCanAdminister: boolean
    selectable?: {
        onSelect: (id: string) => void
        isSelected: (id: string) => boolean
    }
    history: H.History
    location: H.Location
    extensionInfo?: {
        hoverifier: Hoverifier<RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec, HoverMerged, ActionItemAction>
    } & ExtensionsControllerProps
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
    /** For testing only. */
    expandByDefault?: boolean
}

export const ExternalChangesetNode: React.FunctionComponent<ExternalChangesetNodeProps> = ({
    node: initialNode,
    viewerCanAdminister,
    selectable,
    isLightTheme,
    history,
    location,
    extensionInfo,
    queryExternalChangesetWithFileDiffs,
    expandByDefault,
}) => {
    const [node, setNode] = useState(initialNode)
    useEffect(() => {
        setNode(initialNode)
    }, [initialNode])
    const [isExpanded, setIsExpanded] = useState(expandByDefault ?? false)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        event => {
            event.preventDefault()
            setIsExpanded(!isExpanded)
        },
        [isExpanded]
    )

    const selected = selectable?.isSelected(node.id)
    const toggleSelected = useCallback((): void => {
        selectable?.onSelect(node.id)
    }, [selectable, node.id])

    return (
        <>
            <Button
                variant="icon"
                className="test-batches-expand-changeset d-none d-sm-block"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                {isExpanded ? (
                    <ChevronDownIcon className="icon-inline" aria-label="Close section" />
                ) : (
                    <ChevronRightIcon className="icon-inline" aria-label="Expand section" />
                )}
            </Button>
            {selectable ? (
                <div className="p-2">
                    <InputTooltip
                        id={`select-changeset-${node.id}`}
                        type="checkbox"
                        checked={selected}
                        onChange={toggleSelected}
                        disabled={!viewerCanAdminister}
                        tooltip={
                            viewerCanAdminister
                                ? 'Click to select changeset for bulk operation'
                                : 'You do not have permission to perform this operation'
                        }
                    />
                </div>
            ) : (
                // 0-width empty element to allow us to keep the identical grid template of the parent
                // list, regardless of whether or not the nodes have the checkbox selector
                <span />
            )}
            <ChangesetStatusCell
                id={node.id}
                state={node.state}
                className={classNames(
                    styles.externalChangesetNodeState,
                    'p-2 align-self-stretch text-muted d-block d-sm-flex'
                )}
            />
            <ExternalChangesetInfoCell
                node={node}
                viewerCanAdminister={viewerCanAdminister}
                className={classNames(styles.externalChangesetNodeInformation, 'p-2 align-self-stretch')}
            />
            <div
                className={classNames(
                    styles.externalChangesetNodeStatuses,
                    'd-flex d-md-none justify-content-start',
                    (node.checkState || node.reviewState || node.diffStat) && 'p-2'
                )}
            >
                {node.checkState && <ChangesetCheckStatusCell checkState={node.checkState} className="mr-3" />}
                {node.reviewState && <ChangesetReviewStatusCell reviewState={node.reviewState} className="mr-3" />}
                {node.diffStat && <DiffStatStack {...node.diffStat} />}
            </div>
            <span
                className={classNames(
                    'align-self-stretch d-none d-md-flex justify-content-center',
                    node.checkState && 'p-2'
                )}
            >
                {node.checkState && <ChangesetCheckStatusCell checkState={node.checkState} />}
            </span>
            <span
                className={classNames(
                    'align-self-stretch d-none d-md-flex justify-content-center',
                    node.reviewState && 'p-2'
                )}
            >
                {node.reviewState && <ChangesetReviewStatusCell reviewState={node.reviewState} />}
            </span>
            <div
                className={classNames(
                    'align-self-center d-none d-md-flex justify-content-center',
                    node.diffStat && 'p-2'
                )}
            >
                {node.diffStat && <DiffStatStack {...node.diffStat} />}
            </div>
            {/* The button for expanding the information used on xs devices. */}
            <Button
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
                className={classNames(
                    styles.externalChangesetNodeShowDetails,
                    'd-block d-sm-none test-batches-expand-changeset'
                )}
                outline={true}
                variant="secondary"
            >
                {isExpanded ? (
                    <ChevronDownIcon className="icon-inline" aria-label="Close section" />
                ) : (
                    <ChevronRightIcon className="icon-inline" aria-label="Expand section" />
                )}{' '}
                {isExpanded ? 'Hide' : 'Show'} details
            </Button>
            {isExpanded && (
                <>
                    <div className={classNames(styles.externalChangesetNodeBgExpanded, 'align-self-stretch')} />
                    <div
                        className={classNames(
                            styles.externalChangesetNodeExpandedSection,
                            styles.externalChangesetNodeBgExpanded,
                            'p-2'
                        )}
                    >
                        <div className="d-flex justify-content-end">
                            {viewerCanAdminister && node.state === ChangesetState.FAILED && node.error && (
                                <RetryChangesetButton
                                    node={node}
                                    setNode={setNode}
                                    viewerCanAdminister={viewerCanAdminister}
                                />
                            )}
                            {node.currentSpec?.type === ChangesetSpecType.BRANCH && (
                                <DownloadDiffButton changesetID={node.id} />
                            )}
                        </div>
                        {node.syncerError && <SyncerError syncerError={node.syncerError} />}
                        <ChangesetError node={node} />
                        <ChangesetFileDiff
                            changesetID={node.id}
                            isLightTheme={isLightTheme}
                            history={history}
                            location={location}
                            repositoryID={node.repository.id}
                            repositoryName={node.repository.name}
                            extensionInfo={extensionInfo}
                            queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                            updateOnChange={node.updatedAt}
                        />
                    </div>
                </>
            )}
        </>
    )
}

const SyncerError: React.FunctionComponent<{ syncerError: string }> = ({ syncerError }) => (
    <Alert role="alert" variant="danger">
        <h4 className={classNames(styles.alertHeading)}>
            Encountered error during last attempt to sync changeset data from code host
        </h4>
        <ErrorMessage error={syncerError} />
        <hr className="my-2" />
        <p className="mb-0">
            <small>This might be an ephemeral error that resolves itself at the next sync.</small>
        </p>
    </Alert>
)

const ChangesetError: React.FunctionComponent<{
    node: ExternalChangesetFields
}> = ({ node }) => {
    if (!node.error) {
        return null
    }

    return (
        <Alert role="alert" variant="danger">
            <h4 className={classNames(styles.alertHeading)}>Failed to run operations on changeset</h4>
            <ErrorMessage error={node.error} />
        </Alert>
    )
}

const RetryChangesetButton: React.FunctionComponent<{
    node: ExternalChangesetFields
    setNode: (node: ExternalChangesetFields) => void
    viewerCanAdminister: boolean
}> = ({ node, setNode }) => {
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const onRetry = useCallback(async () => {
        setIsLoading(true)
        try {
            const changeset = await reenqueueChangeset(node.id)
            // If repository permissions changed in between - ignore and await fetch (at most 5s) to reflect the new state.
            if (changeset.__typename === 'ExternalChangeset') {
                setIsLoading(false)
                setNode(changeset)
            }
        } catch (error) {
            setIsLoading(asError(error))
        }
    }, [node.id, setNode])
    return (
        <>
            {isErrorLike(isLoading) && <ErrorAlert error={isLoading} prefix="Error re-enqueueing changeset" />}
            <Button className="mb-1" onClick={onRetry} disabled={isLoading === true} variant="link">
                <SyncIcon
                    className={classNames(
                        'icon-inline',
                        isLoading === true && styles.externalChangesetNodeRetrySpinning
                    )}
                />{' '}
                Retry
            </Button>
        </>
    )
}
