import classNames from 'classnames'
import React, { useCallback, useState } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { asError, isErrorLike } from '@sourcegraph/common'
import { Button, LoadingSpinner, Modal, Link } from '@sourcegraph/wildcard'

import { ExternalServiceKind, Scalars } from '../../../graphql-operations'

import styles from './AddCredentialModal.module.scss'
import { createBatchChangesCredential as _createBatchChangesCredential } from './backend'
import { CodeHostSshPublicKey } from './CodeHostSshPublicKey'
import { ModalHeader } from './ModalHeader'

export interface AddCredentialModalProps {
    onCancel: () => void
    afterCreate: () => void
    userID: Scalars['ID'] | null
    externalServiceKind: ExternalServiceKind
    externalServiceURL: string
    requiresSSH: boolean

    /** For testing only. */
    createBatchChangesCredential?: typeof _createBatchChangesCredential
    /** For testing only. */
    initialStep?: Step
}

const HELP_TEXT_LINK_URL = 'https://docs.sourcegraph.com/batch_changes/quickstart#configure-code-host-credentials'

const helpTexts: Record<ExternalServiceKind, JSX.Element> = {
    [ExternalServiceKind.GITHUB]: (
        <>
            <Link to={HELP_TEXT_LINK_URL} rel="noreferrer noopener" target="_blank">
                Create a new access token
            </Link>{' '}
            with the <code>repo</code>, <code>read:org</code>, <code>user:email</code>, <code>read:discussion</code>,
            and <code>workflow</code> scopes.
        </>
    ),
    [ExternalServiceKind.GITLAB]: (
        <>
            <Link to={HELP_TEXT_LINK_URL} rel="noreferrer noopener" target="_blank">
                Create a new access token
            </Link>{' '}
            with <code>api</code>, <code>read_repository</code>, and <code>write_repository</code> scopes.
        </>
    ),
    [ExternalServiceKind.BITBUCKETSERVER]: (
        <>
            <Link to={HELP_TEXT_LINK_URL} rel="noreferrer noopener" target="_blank">
                Create a new access token
            </Link>{' '}
            with <code>write</code> permissions on the project and repository level.
        </>
    ),

    // These are just for type completeness and serve as placeholders for a bright future.
    [ExternalServiceKind.BITBUCKETCLOUD]: <span>Unsupported</span>,
    [ExternalServiceKind.GITOLITE]: <span>Unsupported</span>,
    [ExternalServiceKind.JVMPACKAGES]: <span>Unsupported</span>,
    [ExternalServiceKind.NPMPACKAGES]: <span>Unsupported</span>,
    [ExternalServiceKind.PERFORCE]: <span>Unsupported</span>,
    [ExternalServiceKind.PHABRICATOR]: <span>Unsupported</span>,
    [ExternalServiceKind.AWSCODECOMMIT]: <span>Unsupported</span>,
    [ExternalServiceKind.PAGURE]: <span>Unsupported</span>,
    [ExternalServiceKind.OTHER]: <span>Unsupported</span>,
}

type Step = 'add-token' | 'get-ssh-key'

export const AddCredentialModal: React.FunctionComponent<AddCredentialModalProps> = ({
    onCancel,
    afterCreate,
    userID,
    externalServiceKind,
    externalServiceURL,
    requiresSSH,
    createBatchChangesCredential = _createBatchChangesCredential,
    initialStep = 'add-token',
}) => {
    const labelId = 'addCredential'
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const [credential, setCredential] = useState<string>('')
    const [sshPublicKey, setSSHPublicKey] = useState<string>()
    const [step, setStep] = useState<Step>(initialStep)

    const onChangeCredential = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setCredential(event.target.value)
    }, [])

    const onSubmit = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()
            setIsLoading(true)
            try {
                const createdCredential = await createBatchChangesCredential({
                    user: userID,
                    credential,
                    externalServiceKind,
                    externalServiceURL,
                })
                if (requiresSSH && createdCredential.sshPublicKey) {
                    setSSHPublicKey(createdCredential.sshPublicKey)
                    setStep('get-ssh-key')
                } else {
                    afterCreate()
                }
            } catch (error) {
                setIsLoading(asError(error))
            }
        },
        [
            afterCreate,
            userID,
            credential,
            externalServiceKind,
            externalServiceURL,
            requiresSSH,
            createBatchChangesCredential,
        ]
    )

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <div className="test-add-credential-modal">
                <ModalHeader
                    id={labelId}
                    externalServiceKind={externalServiceKind}
                    externalServiceURL={externalServiceURL}
                />
                {requiresSSH && (
                    <div className="d-flex w-100 justify-content-between mb-4">
                        <div className="flex-grow-1 mr-2">
                            <p className={classNames('mb-0 py-2', step === 'get-ssh-key' && 'text-muted')}>
                                1. Add token
                            </p>
                            <div
                                className={classNames(
                                    styles.addCredentialModalModalStepRuler,
                                    styles.addCredentialModalModalStepRulerPurple
                                )}
                            />
                        </div>
                        <div className="flex-grow-1 ml-2">
                            <p className={classNames('mb-0 py-2', step === 'add-token' && 'text-muted')}>
                                2. Get SSH Key
                            </p>
                            <div
                                className={classNames(
                                    styles.addCredentialModalModalStepRuler,
                                    step === 'add-token' && styles.addCredentialModalModalStepRulerGray,
                                    step === 'get-ssh-key' && styles.addCredentialModalModalStepRulerBlue
                                )}
                            />
                        </div>
                    </div>
                )}
                {step === 'add-token' && (
                    <>
                        {isErrorLike(isLoading) && <ErrorAlert error={isLoading} />}
                        <Form onSubmit={onSubmit}>
                            <div className="form-group">
                                <label htmlFor="token">Personal access token</label>
                                <input
                                    id="token"
                                    name="token"
                                    type="password"
                                    autoComplete="off"
                                    className="form-control test-add-credential-modal-input"
                                    required={true}
                                    spellCheck="false"
                                    minLength={1}
                                    value={credential}
                                    onChange={onChangeCredential}
                                />
                                <p className="form-text">{helpTexts[externalServiceKind]}</p>
                            </div>
                            <div className="d-flex justify-content-end">
                                <Button
                                    disabled={isLoading === true}
                                    className="mr-2"
                                    onClick={onCancel}
                                    outline={true}
                                    variant="secondary"
                                >
                                    Cancel
                                </Button>
                                <Button
                                    type="submit"
                                    disabled={isLoading === true || credential.length === 0}
                                    className="test-add-credential-modal-submit"
                                    variant="primary"
                                >
                                    {isLoading === true && <LoadingSpinner />}
                                    {requiresSSH ? 'Next' : 'Add credential'}
                                </Button>
                            </div>
                        </Form>
                    </>
                )}
                {step === 'get-ssh-key' && (
                    <>
                        <p>
                            An SSH key has been generated for your batch changes code host connection. Copy the public
                            key below and enter it on your code host.
                        </p>
                        <CodeHostSshPublicKey externalServiceKind={externalServiceKind} sshPublicKey={sshPublicKey!} />
                        <div className="d-flex justify-content-end">
                            <Button className="mr-2" onClick={afterCreate} outline={true} variant="secondary">
                                Close
                            </Button>
                            <Button
                                className="test-add-credential-modal-submit"
                                onClick={afterCreate}
                                variant="primary"
                            >
                                Add credential
                            </Button>
                        </div>
                    </>
                )}
            </div>
        </Modal>
    )
}
