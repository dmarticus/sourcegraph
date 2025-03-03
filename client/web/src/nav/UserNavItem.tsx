import { Shortcut } from '@slimsag/react-shortcuts'
import classNames from 'classnames'
import * as H from 'history'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle, Tooltip } from 'reactstrap'

import { KeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts'
import { KEYBOARD_SHORTCUT_SHOW_HELP } from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useTimeoutManager, Link, AnchorLink } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { ThemePreference } from '../stores/themeState'
import { ThemePreferenceProps } from '../theme'
import { UserAvatar } from '../user/UserAvatar'

import styles from './UserNavItem.module.scss'

export interface UserNavItemProps extends ThemeProps, ThemePreferenceProps, ExtensionAlertAnimationProps {
    location: H.Location
    authenticatedUser: Pick<
        AuthenticatedUser,
        'username' | 'avatarURL' | 'settingsURL' | 'organizations' | 'siteAdmin' | 'session' | 'displayName'
    >
    showDotComMarketing: boolean
    keyboardShortcutForSwitchTheme?: KeyboardShortcut
    testIsOpen?: boolean
    codeHostIntegrationMessaging: 'browser-extension' | 'native-integration'
    showRepositorySection?: boolean
}

export interface ExtensionAlertAnimationProps {
    isExtensionAlertAnimating: boolean
}

/**
 * React hook to manage the animation that occurs after the user dismisses
 * `InstallBrowserExtensionAlert`.
 *
 * This hook is called from the the LCA of `UserNavItem` and the component that triggers
 * the animation.
 */
export function useExtensionAlertAnimation(): ExtensionAlertAnimationProps & {
    startExtensionAlertAnimation: () => void
} {
    const [isAnimating, setIsAnimating] = useState(false)

    const animationManager = useTimeoutManager()

    const startExtensionAlertAnimation = useCallback(() => {
        if (!isAnimating) {
            setIsAnimating(true)

            animationManager.setTimeout(() => {
                setIsAnimating(false)
            }, 5100)
        }
    }, [isAnimating, animationManager])

    return { isExtensionAlertAnimating: isAnimating, startExtensionAlertAnimation }
}

/**
 * Triggers Keyboard Shortcut help when the button is clicked in the Menu Nav item
 */

const showKeyboardShortcutsHelp = (): void => {
    const keybinding = KEYBOARD_SHORTCUT_SHOW_HELP.keybindings[0]
    const shiftKey = !!keybinding.held?.includes('Shift')
    const altKey = !!keybinding.held?.includes('Alt')
    const metaKey = !!keybinding.held?.includes('Meta')
    const ctrlKey = !!keybinding.held?.includes('Control')

    for (const key of keybinding.ordered) {
        document.dispatchEvent(new KeyboardEvent('keydown', { key, shiftKey, metaKey, ctrlKey, altKey }))
    }
}

/**
 * Displays the user's avatar and/or username in the navbar and exposes a dropdown menu with more options for
 * authenticated viewers.
 */
export const UserNavItem: React.FunctionComponent<UserNavItemProps> = props => {
    const {
        location,
        themePreference,
        onThemePreferenceChange,
        isExtensionAlertAnimating,
        testIsOpen,
        codeHostIntegrationMessaging,
    } = props

    const supportsSystemTheme = useMemo(
        () => Boolean(window.matchMedia?.('not all and (prefers-color-scheme), (prefers-color-scheme)').matches),
        []
    )

    const [isOpen, setIsOpen] = useState(() => !!testIsOpen)
    const toggleIsOpen = useCallback(() => setIsOpen(open => !open), [])

    useEffect(() => {
        // Close dropdown after clicking on a dropdown item.
        if (!testIsOpen) {
            setIsOpen(false)
        }
    }, [location.pathname, testIsOpen])

    const onThemeChange: React.ChangeEventHandler<HTMLSelectElement> = useCallback(
        event => {
            onThemePreferenceChange(event.target.value as ThemePreference)
        },
        [onThemePreferenceChange]
    )

    const onThemeCycle = useCallback((): void => {
        onThemePreferenceChange(themePreference === ThemePreference.Dark ? ThemePreference.Light : ThemePreference.Dark)
    }, [onThemePreferenceChange, themePreference])

    // Target ID for tooltip
    const targetID = 'target-user-avatar'

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className="py-0" aria-label="User. Open menu">
            <DropdownToggle className="bg-transparent d-flex align-items-center test-user-nav-item-toggle" nav={true}>
                <div className="position-relative">
                    <div className="align-items-center d-flex">
                        <UserAvatar
                            user={props.authenticatedUser}
                            targetID={targetID}
                            className={classNames('icon-inline', styles.avatar)}
                        />
                        {isOpen ? (
                            <ChevronUpIcon className="icon-inline" />
                        ) : (
                            <ChevronDownIcon className="icon-inline" />
                        )}
                    </div>
                </div>
                {isExtensionAlertAnimating && (
                    <Tooltip
                        target={targetID}
                        placement="bottom"
                        isOpen={true}
                        modifiers={{
                            offset: {
                                offset: '0, 10px',
                            },
                        }}
                        className={styles.tooltip}
                    >
                        Install the browser extension from here later
                    </Tooltip>
                )}
            </DropdownToggle>
            <DropdownMenu right={true} className={styles.dropdownMenu}>
                <DropdownItem header={true} className="py-1">
                    Signed in as <strong>@{props.authenticatedUser.username}</strong>
                </DropdownItem>
                <DropdownItem divider={true} />
                <Link to={props.authenticatedUser.settingsURL!} className="dropdown-item">
                    Settings
                </Link>
                {props.showRepositorySection && (
                    <Link
                        to={`/users/${props.authenticatedUser.username}/settings/repositories`}
                        className="dropdown-item"
                    >
                        Your repositories
                    </Link>
                )}
                <Link to={`/users/${props.authenticatedUser.username}/searches`} className="dropdown-item">
                    Saved searches
                </Link>
                <DropdownItem divider={true} />
                <div className="px-2 py-1">
                    <div className="d-flex align-items-center">
                        <div className="mr-2">Theme</div>
                        <select
                            className="custom-select custom-select-sm test-theme-toggle"
                            onChange={onThemeChange}
                            value={props.themePreference}
                        >
                            <option value={ThemePreference.Light}>Light</option>
                            <option value={ThemePreference.Dark}>Dark</option>
                            <option value={ThemePreference.System}>System</option>
                        </select>
                    </div>
                    {props.themePreference === ThemePreference.System && !supportsSystemTheme && (
                        <div className="text-wrap">
                            <small>
                                <Link
                                    to="https://caniuse.com/#feat=prefers-color-scheme"
                                    className="text-warning"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    Your browser does not support the system theme.
                                </Link>
                            </small>
                        </div>
                    )}
                    {props.keyboardShortcutForSwitchTheme?.keybindings.map((keybinding, index) => (
                        <Shortcut key={index} {...keybinding} onMatch={onThemeCycle} />
                    ))}
                </div>
                {props.authenticatedUser.organizations.nodes.length > 0 && (
                    <>
                        <DropdownItem divider={true} />
                        <DropdownItem header={true}>Your organizations</DropdownItem>
                        {props.authenticatedUser.organizations.nodes.map(org => (
                            <Link key={org.id} to={org.settingsURL || org.url} className="dropdown-item">
                                {org.displayName || org.name}
                            </Link>
                        ))}
                    </>
                )}
                <DropdownItem divider={true} />
                {props.authenticatedUser.siteAdmin && (
                    <Link to="/site-admin" className="dropdown-item">
                        Site admin
                    </Link>
                )}
                <Link to="/help" className="dropdown-item" target="_blank" rel="noopener">
                    Help <OpenInNewIcon className="icon-inline" />
                </Link>
                <button onClick={showKeyboardShortcutsHelp} type="button" className="dropdown-item">
                    Keyboard shortcuts
                </button>

                {props.authenticatedUser.session?.canSignOut && (
                    <AnchorLink to="/-/sign-out" className="dropdown-item">
                        Sign out
                    </AnchorLink>
                )}
                <DropdownItem divider={true} />
                {props.showDotComMarketing && (
                    <Link to="https://about.sourcegraph.com" target="_blank" rel="noopener" className="dropdown-item">
                        About Sourcegraph <OpenInNewIcon className="icon-inline" />
                    </Link>
                )}
                {codeHostIntegrationMessaging === 'browser-extension' && (
                    <Link
                        to="https://docs.sourcegraph.com/integration/browser_extension"
                        target="_blank"
                        rel="noopener"
                        className="dropdown-item"
                    >
                        Browser extension <OpenInNewIcon className="icon-inline" />
                    </Link>
                )}
            </DropdownMenu>
        </ButtonDropdown>
    )
}
