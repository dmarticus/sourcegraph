import classNames from 'classnames'
import ExportIcon from 'mdi-react/ExportIcon'
import PlusThickIcon from 'mdi-react/PlusThickIcon'
import React, { useMemo } from 'react'
import FocusLock from 'react-focus-lock'
import { Popover } from 'reactstrap'

import { ExternalServiceKind } from '@sourcegraph/shared/src/schema'
import { ButtonLink } from '@sourcegraph/wildcard'

import { SourcegraphIcon } from '../../auth/icons'

import { serviceKindDisplayNameAndIcon } from './GoToCodeHostAction'
import styles from './InstallBrowserExtensionPopover.module.scss'

interface Props {
    url: string
    serviceKind: ExternalServiceKind | null
    onClose: () => void
    onReject: () => void
    onInstall: () => void
    targetID: string
    onToggle: () => void
    isOpen: boolean
}

export const InstallBrowserExtensionPopover: React.FunctionComponent<Props> = ({
    url,
    serviceKind,
    onClose,
    onReject,
    onInstall,
    targetID,
    onToggle,
    isOpen,
}) => {
    const { displayName, icon } = serviceKindDisplayNameAndIcon(serviceKind)
    const Icon = icon || ExportIcon

    // Open all external links in new tab
    const linkProps = { rel: 'noopener noreferrer', target: '_blank' }

    return (
        <Popover
            toggle={onToggle}
            target={targetID}
            isOpen={isOpen}
            popperClassName={classNames('shadow border', styles.installBrowserExtensionPopover)}
            innerClassName="border-0"
            placement="bottom"
            boundariesElement="window"
            modifiers={useMemo(
                () => ({
                    offset: {
                        offset: '0, 0',
                        enabled: true,
                    },
                }),
                []
            )}
        >
            {isOpen && (
                <FocusLock returnFocus={true}>
                    <div className="p-3 text-wrap  test-install-browser-extension-popover">
                        <h3 className="mb-0 test-install-browser-extension-popover-header">
                            Take Sourcegraph's code intelligence to {displayName}!
                        </h3>
                        <p className="py-3">
                            Install Sourcegraph browser extension to add code intelligence{' '}
                            {serviceKind === ExternalServiceKind.PHABRICATOR
                                ? 'while browsing and reviewing code'
                                : `to ${serviceKind === ExternalServiceKind.GITLAB ? 'MR' : 'PR'}s and file views`}{' '}
                            on {displayName} or any other connected code host.
                        </p>

                        <div
                            className={classNames(
                                'mx-auto d-flex justify-content-between align-items-center',
                                styles.graphicContainer
                            )}
                        >
                            <SourcegraphIcon className={classNames('p-1', styles.logo)} />
                            <PlusThickIcon className={styles.plusIcon} />
                            <Icon className={styles.logo} />
                        </div>

                        <div className="d-flex justify-content-end">
                            <ButtonLink
                                className="mr-2"
                                onSelect={onReject}
                                to={url}
                                {...linkProps}
                                variant="secondary"
                                outline={true}
                            >
                                No, thanks
                            </ButtonLink>

                            <ButtonLink
                                className="mr-2"
                                onSelect={onClose}
                                to={url}
                                {...linkProps}
                                variant="secondary"
                                outline={true}
                            >
                                Remind me later
                            </ButtonLink>

                            <ButtonLink
                                className="mr-2"
                                onSelect={onInstall}
                                to="/help/integration/browser_extension"
                                {...linkProps}
                                variant="primary"
                            >
                                Install browser extension
                            </ButtonLink>
                        </div>
                    </div>
                </FocusLock>
            )}
        </Popover>
    )
}
