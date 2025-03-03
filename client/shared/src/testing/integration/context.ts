import * as path from 'path'
import * as util from 'util'

import { MODE, Polly, PollyServer } from '@pollyjs/core'
import FSPersister from '@pollyjs/persister-fs'
import { GraphQLError } from 'graphql'
import { snakeCase } from 'lodash'
import * as mime from 'mime-types'
import { Test } from 'mocha'
import { readFile, mkdir } from 'mz/fs'
import pTimeout from 'p-timeout'
import * as prettier from 'prettier'
import { Subject, Subscription, throwError } from 'rxjs'
import { first, timeoutWith } from 'rxjs/operators'

import { asError, keyExistsIn } from '@sourcegraph/common'
import { ErrorGraphQLResult, SuccessGraphQLResult } from '@sourcegraph/http-client'

import { recordCoverage } from '../coverage'
import { Driver } from '../driver'
import { readEnvironmentString } from '../utils'

import { CdpAdapter, CdpAdapterOptions } from './polly/CdpAdapter'

// Reduce log verbosity
util.inspect.defaultOptions.depth = 0
util.inspect.defaultOptions.maxStringLength = 80

Polly.register(CdpAdapter as any)
Polly.register(FSPersister)

const ASSETS_DIRECTORY = path.resolve(__dirname, '../../../../../ui/assets')

const checkPollyMode = (mode: string): MODE => {
    if (mode === 'record' || mode === 'replay' || mode === 'passthrough' || mode === 'stopped') {
        return mode
    }

    throw new Error(`Invalid Polly mode (check POLLYJS_MODE): ${mode}`)
}

const pollyMode = checkPollyMode(readEnvironmentString({ variable: 'POLLYJS_MODE', defaultValue: 'passthrough' }))

export class IntegrationTestGraphQlError extends Error {
    constructor(public errors: GraphQLError[]) {
        super('graphql error for integration tests')
    }
}

export interface IntegrationTestContext<
    TGraphQlOperations extends Record<TGraphQlOperationNames, (variables: any) => any>,
    TGraphQlOperationNames extends string
> {
    driver: Driver
    server: PollyServer

    /**
     * Configures fake responses for GraphQL queries and mutations.
     *
     * @param overrides The results to return, keyed by query name.
     */
    overrideGraphQL: (overrides: Partial<TGraphQlOperations>) => void

    /**
     * Waits for a specific GraphQL query to happen and returns the variables passed to the request.
     * If the query does not happen within a few seconds, it throws a timeout error.
     *
     * @param triggerRequest A callback called to trigger the request (e.g. clicking a button). The request MUST be triggered within this callback.
     * @param operationName The name of the query to wait for.
     * @returns The GraphQL variables of the query.
     */
    waitForGraphQLRequest: <O extends TGraphQlOperationNames>(
        triggerRequest: () => Promise<void> | void,
        operationName: O
    ) => Promise<Parameters<TGraphQlOperations[O]>[0]>

    dispose: () => Promise<void>
}

export interface IntegrationTestOptions {
    /**
     * The test driver created in a `before()` hook.
     */
    driver: Driver

    /**
     * The value of `this.currentTest` in the `beforeEach()` hook.
     * Make sure the hook function is not an arrow function to access it.
     */
    currentTest: Test

    /**
     * The directory (value of `__dirname`) of the test file.
     */
    directory: string
}

const DISPOSE_ACTION_TIMEOUT = 5 * 1000

/**
 * Should be called in a `beforeEach()` and saved into a local variable.
 */
export const createSharedIntegrationTestContext = async <
    TGraphQlOperations extends Record<TGraphQlOperationNames, (variables: any) => any>,
    TGraphQlOperationNames extends string
>({
    driver,
    currentTest,
    directory,
}: IntegrationTestOptions): Promise<IntegrationTestContext<TGraphQlOperations, TGraphQlOperationNames>> => {
    await driver.newPage()
    const recordingsDirectory = path.join(directory, '__fixtures__', snakeCase(currentTest.fullTitle()))
    if (pollyMode === 'record') {
        await mkdir(recordingsDirectory, { recursive: true })
    }
    const subscriptions = new Subscription()
    const cdpAdapterOptions: CdpAdapterOptions = {
        browser: driver.browser,
    }
    const polly = new Polly(snakeCase(currentTest.title), {
        adapters: [CdpAdapter.id],
        adapterOptions: {
            [CdpAdapter.id]: cdpAdapterOptions,
        },
        persister: FSPersister.id,
        persisterOptions: {
            [FSPersister.id]: {
                recordingsDir: recordingsDirectory,
            },
        },
        expiryStrategy: 'warn',
        recordIfMissing: pollyMode === 'record',
        matchRequestsBy: {
            method: true,
            body: true,
            order: true,
            // Origin header will change when running against a test instance
            headers: false,
        },
        mode: pollyMode,
        logging: false,
    })
    const { server } = polly

    // Fail the test in the case a request handler threw an error,
    // e.g. because a request had no mock defined.
    const cdpAdapter = polly.adapters.get(CdpAdapter.id) as CdpAdapter
    subscriptions.add(
        cdpAdapter.errors.subscribe(error => {
            currentTest.emit('error', error)
        })
    )

    // Let browser handle data: URIs
    server.get('data:*rest').passthrough()

    // Special URL: The browser redirects to chrome-extension://invalid
    // when requesting an extension resource that does not exist.
    server.get('chrome-extension://invalid/').passthrough()

    // Avoid 404 error logs from missing favicon
    server.get(new URL('/favicon.ico', driver.sourcegraphBaseUrl).href).intercept((request, response) => {
        response
            .status(302)
            .setHeader('Location', new URL('/.assets/img/sourcegraph-mark.svg', driver.sourcegraphBaseUrl).href)
            .send('')
    })

    // Serve assets from disk
    server.get(new URL('/.assets/*path', driver.sourcegraphBaseUrl).href).intercept(async (request, response) => {
        const asset = request.params.path
        // Cache all responses for the entire lifetime of the test run
        response.setHeader('Cache-Control', 'public, max-age=31536000, immutable')
        try {
            const content = await readFile(path.join(ASSETS_DIRECTORY, asset), {
                // Polly doesn't support Buffers or streams at the moment
                encoding: 'utf-8',
            })
            const contentType = mime.contentType(path.basename(asset))
            if (contentType) {
                response.type(contentType)
            }
            response.send(content)
        } catch (error) {
            if ((asError(error) as NodeJS.ErrnoException).code === 'ENOENT') {
                response.sendStatus(404)
            } else {
                console.error(error)
                response.status(500).send(asError(error).message)
            }
        }
    })

    // GraphQL requests are not handled by HARs, but configured per-test.
    interface GraphQLRequestEvent<O extends TGraphQlOperationNames> {
        operationName: O
        variables: Parameters<TGraphQlOperations[O]>[0]
    }
    let graphQlOverrides: Partial<TGraphQlOperations> = {}
    const graphQlRequests = new Subject<GraphQLRequestEvent<TGraphQlOperationNames>>()
    server.post(new URL('/.api/graphql', driver.sourcegraphBaseUrl).href).intercept((request, response) => {
        const operationName = new URL(request.absoluteUrl).search.slice(1) as TGraphQlOperationNames
        const { variables, query } = request.jsonBody() as {
            query: string
            variables: Parameters<TGraphQlOperations[TGraphQlOperationNames]>[0]
        }
        graphQlRequests.next({ operationName, variables })

        const missingOverrideError = (): Error => {
            const formattedQuery = prettier.format(query, { parser: 'graphql' }).trim()
            const formattedVariables = util.inspect(variables)
            const error = new Error(
                `GraphQL query "${operationName}" has no configured mock response. Make sure the call to overrideGraphQL() includes a result for the "${operationName}" query:\n${formattedVariables} ⤵️\n${formattedQuery}`
            )
            return error
        }
        if (!graphQlOverrides || !keyExistsIn(operationName, graphQlOverrides)) {
            throw missingOverrideError()
        }
        const handler = graphQlOverrides[operationName]
        if (!handler) {
            throw missingOverrideError()
        }

        try {
            const result = handler(variables as any)
            const graphQlResult: SuccessGraphQLResult<any> = { data: result, errors: undefined }
            response.json(graphQlResult)
        } catch (error) {
            if (!(error instanceof IntegrationTestGraphQlError)) {
                throw error
            }

            const graphQlError: ErrorGraphQLResult = { data: undefined, errors: error.errors }
            response.json(graphQlError)
        }
    })

    // Filter out 'server' header filled in by Caddy before persisting responses,
    // otherwise tests will hang when replayed from recordings.
    server
        .any()
        .on('beforePersist', (request, recording: { response: { headers: { name: string; value: string }[] } }) => {
            recording.response.headers = recording.response.headers.filter(({ name }) => name !== 'server')
        })

    return {
        driver,
        server,
        overrideGraphQL: overrides => {
            graphQlOverrides = { ...graphQlOverrides, ...overrides }
        },
        waitForGraphQLRequest: async <O extends TGraphQlOperationNames>(
            triggerRequest: () => Promise<void> | void,
            operationName: O
        ): Promise<Parameters<TGraphQlOperations[O]>[0]> => {
            const requestPromise = graphQlRequests
                .pipe(
                    first(
                        (request: GraphQLRequestEvent<TGraphQlOperationNames>): request is GraphQLRequestEvent<O> =>
                            request.operationName === operationName
                    ),
                    timeoutWith(4000, throwError(new Error(`Timeout waiting for GraphQL request "${operationName}"`)))
                )
                .toPromise()
            await triggerRequest()
            const { variables } = await requestPromise
            return variables
        },
        dispose: async () => {
            subscriptions.unsubscribe()
            await pTimeout(
                recordCoverage(driver.browser),
                DISPOSE_ACTION_TIMEOUT,
                new Error('Recording coverage timed out')
            )

            if (driver.page.url() !== 'about:blank') {
                await pTimeout(
                    driver.page.evaluate(() => {
                        try {
                            localStorage.clear()
                        } catch (error) {
                            console.error('Failed to clear localStorage!', error)
                        }
                    }),
                    DISPOSE_ACTION_TIMEOUT,
                    () => console.warn('Failed to clear localStorage!')
                )
            }

            await pTimeout(driver.page.close(), DISPOSE_ACTION_TIMEOUT, new Error('Closing Puppeteer page timed out'))
            await pTimeout(polly.stop(), DISPOSE_ACTION_TIMEOUT, new Error('Stopping Polly timed out'))
        },
    }
}
