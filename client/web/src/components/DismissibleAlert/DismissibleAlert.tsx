import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'

import { Button, Alert, AlertProps } from '@sourcegraph/wildcard'

import styles from './DismissibleAlert.module.scss'

export interface DismissibleAlertProps extends AlertProps {
    /**
     * If provided, used to build the key that represents the alert in local storage. An
     * alert with a storage key will be permanently dismissed once the user dismisses it.
     */
    partialStorageKey?: string

    testId?: string
}

/**
 * A global site alert that can be dismissed. If a `partialStorageKey` is provided, the
 * alert will never be shown again after it is dismissed. Otherwise, it will be shown
 * whenever unmounted and remounted.
 */
export const DismissibleAlert: React.FunctionComponent<DismissibleAlertProps> = ({
    partialStorageKey,
    className,
    testId,
    children,
    variant,
}) => {
    const [dismissed, setDismissed] = React.useState<boolean>(
        partialStorageKey ? isAlertDismissed(partialStorageKey) : false
    )

    const onDismiss = React.useCallback(() => {
        if (partialStorageKey) {
            dismissAlert(partialStorageKey)
        }
        setDismissed(true)
    }, [partialStorageKey])

    if (dismissed) {
        return null
    }

    return (
        <Alert data-testid={testId} className={classNames(styles.container, className)} variant={variant}>
            <div className={styles.content}>{children}</div>
            <Button aria-label="Close alert" variant="icon" className={styles.closeButton} onClick={onDismiss}>
                <CloseIcon className="icon-inline" />
            </Button>
        </Alert>
    )
}

export function dismissAlert(key: string): void {
    localStorage.setItem(storageKeyForPartial(key), 'true')
}

export function isAlertDismissed(key: string): boolean {
    return localStorage.getItem(storageKeyForPartial(key)) === 'true'
}

function storageKeyForPartial(partialStorageKey: string): string {
    return `DismissibleAlert/${partialStorageKey}/dismissed`
}
