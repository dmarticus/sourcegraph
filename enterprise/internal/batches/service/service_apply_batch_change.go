package service

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/rewirer"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ErrApplyClosedBatchChange is returned by ApplyBatchChange when the batch change
// matched by the batch spec is already closed.
var ErrApplyClosedBatchChange = errors.New("existing batch change matched by batch spec is closed")

// ErrMatchingBatchChangeExists is returned by ApplyBatchChange if a batch change matching the
// batch spec already exists and FailIfExists was set.
var ErrMatchingBatchChangeExists = errors.New("a batch change matching the given batch spec already exists")

// ErrEnsureBatchChangeFailed is returned by AppplyBatchChange when a
// ensureBatchChangeID is provided but a batch change with the name specified the
// batchSpec exists in the given namespace but has a different ID.
var ErrEnsureBatchChangeFailed = errors.New("a batch change in the given namespace and with the given name exists but does not match the given ID")

type ApplyBatchChangeOpts struct {
	BatchSpecRandID     string
	EnsureBatchChangeID int64

	// When FailIfBatchChangeExists is true, ApplyBatchChange will fail if a batch change
	// matching the given batch spec already exists.
	FailIfBatchChangeExists bool

	PublicationStates UiPublicationStates
}

func (o ApplyBatchChangeOpts) String() string {
	return fmt.Sprintf(
		"BatchSpec %s, EnsureBatchChangeID %d",
		o.BatchSpecRandID,
		o.EnsureBatchChangeID,
	)
}

// ApplyBatchChange creates the BatchChange.
func (s *Service) ApplyBatchChange(
	ctx context.Context,
	opts ApplyBatchChangeOpts,
) (batchChange *btypes.BatchChange, err error) {
	ctx, endObservation := s.operations.applyBatchChange.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	batchSpec, err := s.store.GetBatchSpec(ctx, store.GetBatchSpecOpts{
		RandID: opts.BatchSpecRandID,
	})
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only site-admins or the creator of batchSpec can apply it.
	if err := backend.CheckSiteAdminOrSameUser(ctx, s.store.DatabaseDB(), batchSpec.UserID); err != nil {
		return nil, err
	}

	// Validate ChangesetSpecs and return error if they're invalid and the
	// BatchSpec can't be applied safely.
	if err := s.ValidateChangesetSpecs(ctx, batchSpec.ID); err != nil {
		return nil, err
	}

	batchChange, previousSpecID, err := s.ReconcileBatchChange(ctx, batchSpec)
	if err != nil {
		return nil, err
	}

	if batchChange.ID != 0 && opts.FailIfBatchChangeExists {
		return nil, ErrMatchingBatchChangeExists
	}

	if opts.EnsureBatchChangeID != 0 && batchChange.ID != opts.EnsureBatchChangeID {
		return nil, ErrEnsureBatchChangeFailed
	}

	if batchChange.Closed() {
		return nil, ErrApplyClosedBatchChange
	}

	if previousSpecID == batchSpec.ID {
		return batchChange, nil
	}

	// Before we write to the database in a transaction, we cancel all
	// currently enqueued/errored-and-retryable changesets the batch change might
	// have.
	// We do this so we don't continue to possibly create changesets on the
	// codehost while we're applying a new batch spec.
	// This is blocking, because the changeset rows currently being processed by the
	// reconciler are locked.
	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	l := locker.NewWithDB(nil, "batches_apply").With(tx)
	locked, err := l.LockInTransaction(ctx, int32(batchChange.ID), false)
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, errors.New("batch change locked by other user applying batch spec")
	}

	if err := tx.CancelQueuedBatchChangeChangesets(ctx, batchChange.ID); err != nil {
		return batchChange, nil
	}

	if batchChange.ID == 0 {
		if err := tx.CreateBatchChange(ctx, batchChange); err != nil {
			return nil, err
		}
	} else {
		if err := tx.UpdateBatchChange(ctx, batchChange); err != nil {
			return nil, err
		}
	}

	// Now we need to wire up the ChangesetSpecs of the new BatchSpec
	// correctly with the Changesets so that the reconciler can create/update
	// them.

	// Load the mapping between ChangesetSpecs and existing Changesets in the target batch spec.
	mappings, err := tx.GetRewirerMappings(ctx, store.GetRewirerMappingsOpts{
		BatchSpecID:   batchChange.BatchSpecID,
		BatchChangeID: batchChange.ID,
	})
	if err != nil {
		return nil, err
	}

	// And execute the mapping.
	changesets, err := rewirer.New(mappings, batchChange.ID).Rewire()
	if err != nil {
		return nil, err
	}

	// Prepare the UI publication states. We need to do this within the
	// transaction to avoid conflicting writes to the changeset specs.
	if err := opts.PublicationStates.prepareAndValidate(mappings); err != nil {
		return nil, err
	}

	// Upsert all changesets.
	for _, changeset := range changesets {
		if state := opts.PublicationStates.get(changeset.CurrentSpecID); state != nil {
			changeset.UiPublicationState = state
		}

		if err := tx.UpsertChangeset(ctx, changeset); err != nil {
			return nil, err
		}
	}

	return batchChange, nil
}

func (s *Service) ReconcileBatchChange(
	ctx context.Context,
	batchSpec *btypes.BatchSpec,
) (batchChange *btypes.BatchChange, previousSpecID int64, err error) {
	ctx, endObservation := s.operations.reconcileBatchChange.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	batchChange, err = s.GetBatchChangeMatchingBatchSpec(ctx, batchSpec)
	if err != nil {
		return nil, 0, err
	}
	if batchChange == nil {
		batchChange = &btypes.BatchChange{}
	} else {
		previousSpecID = batchChange.BatchSpecID
	}
	// Populate the batch change with the values from the batch spec.
	batchChange.BatchSpecID = batchSpec.ID
	batchChange.NamespaceOrgID = batchSpec.NamespaceOrgID
	batchChange.NamespaceUserID = batchSpec.NamespaceUserID
	batchChange.Name = batchSpec.Spec.Name
	a := actor.FromContext(ctx)
	if batchChange.CreatorID == 0 {
		batchChange.CreatorID = a.UID
	}
	batchChange.LastApplierID = a.UID
	batchChange.LastAppliedAt = s.clock()
	batchChange.Description = batchSpec.Spec.Description
	return batchChange, previousSpecID, nil
}
