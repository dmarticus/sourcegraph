import { parseISO } from 'date-fns'
import formatDistanceStrict from 'date-fns/formatDistanceStrict'
import { isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useEffect, useMemo } from 'react'
import { Observable, of } from 'rxjs'
import { catchError, map, startWith } from 'rxjs/operators'

import { asError, createAggregateError, ErrorLike, isErrorLike, numberWithCommas } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import * as GQL from '@sourcegraph/shared/src/schema'
import { LoadingSpinner, useObservable, Alert, Link } from '@sourcegraph/wildcard'

import { queryGraphQL } from '../../../backend/graphql'
import { formatUserCount, mailtoSales } from '../../../productSubscription/helpers'

import { ProductSubscriptionBeforeAfterInvoiceItem } from './ProductSubscriptionBeforeAfterInvoiceItem'
import { PaymentValidity } from './ProductSubscriptionForm'

interface Props {
    /**
     * The ID of the account associated with the subscription, or null if there is none (in which case the
     * subscription price can be quoted but the subscription can't be bought).
     */
    accountID: Scalars['ID'] | null

    /** The existing product subscription to edit, or null if this is a new subscription. */
    subscriptionID: Scalars['ID'] | null

    /**
     * The product subscription chosen by the user, or null for an invalid choice.
     */
    productSubscription: GQL.IProductSubscriptionInput | null

    /**
     * Called when the validity state of the payment and billing information changes. Initially it
     * is always Invalid.
     */
    onValidityChange: (value: PaymentValidity) => void

    /** For mocking in tests only. */
    _queryPreviewProductSubscriptionInvoice?: typeof queryPreviewProductSubscriptionInvoice
}

const LOADING = 'loading' as const

type PreviewInvoiceOrError = GQL.IProductSubscriptionPreviewInvoice | null | typeof LOADING | ErrorLike

const previewInvoiceValidity = (previewInvoice: PreviewInvoiceOrError): PaymentValidity =>
    previewInvoice === null ||
    previewInvoice === LOADING ||
    isErrorLike(previewInvoice) ||
    isEqual(previewInvoice.beforeInvoiceItem, previewInvoice.afterInvoiceItem) ||
    previewInvoice.isDowngradeRequiringManualIntervention
        ? PaymentValidity.Invalid
        : previewInvoice.price === 0
        ? PaymentValidity.NoPaymentRequired
        : PaymentValidity.Valid

const undefinedIsLoading = <T extends any>(value: T | undefined): T | typeof LOADING =>
    value === undefined ? LOADING : value

/**
 * Displays the payment section of the new product subscription form.
 */
export const NewProductSubscriptionPaymentSection: React.FunctionComponent<Props> = ({
    accountID,
    subscriptionID,
    productSubscription,
    onValidityChange,
    _queryPreviewProductSubscriptionInvoice = queryPreviewProductSubscriptionInvoice,
}) => {
    /**
     * The preview invoice for the subscription, null if the input is invalid to generate an
     * invoice, loading, or an error.
     */
    const previewInvoice = undefinedIsLoading<PreviewInvoiceOrError>(
        useObservable(
            useMemo((): Observable<PreviewInvoiceOrError> => {
                if (productSubscription === null) {
                    return of(null)
                }
                return _queryPreviewProductSubscriptionInvoice({
                    account: accountID,
                    subscriptionToUpdate: subscriptionID,
                    productSubscription,
                }).pipe(
                    catchError(error => [asError(error)]),
                    startWith(LOADING)
                )
            }, [_queryPreviewProductSubscriptionInvoice, accountID, productSubscription, subscriptionID])
        )
    )

    useEffect(() => {
        onValidityChange(previewInvoiceValidity(previewInvoice))
    }, [onValidityChange, previewInvoice])

    return (
        <div className="new-product-subscription-payment-section">
            <div className="form-text mb-2">
                {previewInvoice === LOADING ? (
                    <LoadingSpinner />
                ) : !productSubscription || previewInvoice === null ? (
                    <>&mdash;</>
                ) : isErrorLike(previewInvoice) ? (
                    <span className="text-danger">
                        <AlertCircleIcon className="icon-inline" data-tooltip={previewInvoice.message} /> Error
                    </span>
                ) : previewInvoice.beforeInvoiceItem ? (
                    <>
                        <ProductSubscriptionBeforeAfterInvoiceItem
                            beforeInvoiceItem={previewInvoice.beforeInvoiceItem}
                            afterInvoiceItem={previewInvoice.afterInvoiceItem}
                            className="mb-2"
                        />
                        {previewInvoice.isDowngradeRequiringManualIntervention ? (
                            <Alert className="mb-2" variant="danger">
                                Self-service downgrades are not yet supported.{' '}
                                <Link
                                    to={mailtoSales({
                                        subject: `Downgrade subscription ${subscriptionID!}`,
                                    })}
                                >
                                    Contact sales
                                </Link>{' '}
                                for help.
                            </Alert>
                        ) : (
                            !isEqual(previewInvoice.beforeInvoiceItem, previewInvoice.afterInvoiceItem) && (
                                <div className="mb-2">Amount due: ${numberWithCommas(previewInvoice.price / 100)}</div>
                            )
                        )}
                    </>
                ) : (
                    <>
                        Total: ${numberWithCommas(previewInvoice.price / 100)} for{' '}
                        {formatDistanceStrict(parseISO(previewInvoice.afterInvoiceItem.expiresAt), Date.now())} (
                        {formatUserCount(previewInvoice.afterInvoiceItem.userCount)})
                        {/* Include invisible LoadingSpinner to ensure that the height remains constant between loading and total. */}
                        <LoadingSpinner className="invisible" />
                    </>
                )}
            </div>
        </div>
    )
}

function queryPreviewProductSubscriptionInvoice(
    args: GQL.IPreviewProductSubscriptionInvoiceOnDotcomQueryArguments
): Observable<GQL.IProductSubscriptionPreviewInvoice> {
    return queryGraphQL(
        gql`
            query PreviewProductSubscriptionInvoice(
                $account: ID
                $subscriptionToUpdate: ID
                $productSubscription: ProductSubscriptionInput!
            ) {
                dotcom {
                    previewProductSubscriptionInvoice(
                        account: $account
                        subscriptionToUpdate: $subscriptionToUpdate
                        productSubscription: $productSubscription
                    ) {
                        price
                        prorationDate
                        isDowngradeRequiringManualIntervention
                        beforeInvoiceItem {
                            plan {
                                billingPlanID
                                name
                                pricePerUserPerYear
                            }
                            userCount
                            expiresAt
                        }
                        afterInvoiceItem {
                            plan {
                                billingPlanID
                                name
                                pricePerUserPerYear
                            }
                            userCount
                            expiresAt
                        }
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (
                !data ||
                !data.dotcom ||
                !data.dotcom.previewProductSubscriptionInvoice ||
                (errors && errors.length > 0)
            ) {
                throw createAggregateError(errors)
            }
            return data.dotcom.previewProductSubscriptionInvoice
        })
    )
}
