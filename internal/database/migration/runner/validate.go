package runner

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
)

func (r *Runner) Validate(ctx context.Context, schemaNames ...string) error {
	return r.forEachSchema(ctx, schemaNames, func(ctx context.Context, schemaContext schemaContext) error {
		return r.validateSchema(ctx, schemaContext)
	})
}

// validateSchema returns a non-nil error value if the target database schema is not in the state
// expected by the given schema context. This method will block if there are relevant migrations
// in progress.
func (r *Runner) validateSchema(ctx context.Context, schemaContext schemaContext) error {
	definitions := schemaContext.schema.Definitions.All()

	// Quickly determine with our initial schema version if we are up to date. If so, we won't need
	// to take an advisory lock and poll index creation status below.
	byState := groupByState(schemaContext.initialSchemaVersion, definitions)
	if len(byState.pending) == 0 && len(byState.failed) == 0 && len(byState.applied) == len(definitions) {
		return nil
	}

	for {
		// Attempt to validate the given definitions. We may have to call this several times as
		// we are unable to hold a consistent advisory lock in the presence of migrations utilizing
		// concurrent index creation. Therefore, some invocations of this method will return with
		// a flag to request re-invocation under a new lock.

		if retry, err := r.validateDefinitions(ctx, schemaContext, definitions); err != nil {
			return err
		} else if !retry {
			break
		}

		// There are active index creation operations ongoing; wait a short time before requerying
		// the state of the migrations so we don't flood the database with constant queries to the
		// system catalog.
		if err := wait(ctx, indexPollInterval); err != nil {
			return err
		}
	}

	return nil
}

// validateDefinitions attempts to take an advisory lock, then re-checks the version of the database.
// If there are still migrations to apply from the given definitions, an error is returned.
func (r *Runner) validateDefinitions(ctx context.Context, schemaContext schemaContext, definitions []definition.Definition) (retry bool, _ error) {
	return r.withLockedSchemaState(ctx, schemaContext, definitions, func(schemaVersion schemaVersion, byState definitionsByState, _ unlockFunc) error {
		if len(byState.applied) != len(definitions) {
			// Return an error if all expected schemas have not been applied
			return newOutOfDateError(schemaContext, schemaVersion)
		}

		return nil
	})
}
