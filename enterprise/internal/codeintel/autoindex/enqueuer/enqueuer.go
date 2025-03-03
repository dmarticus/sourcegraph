package enqueuer

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"
	"golang.org/x/time/rate"

	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/inference"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type IndexEnqueuer struct {
	dbStore            DBStore
	gitserverClient    GitserverClient
	repoUpdater        RepoUpdaterClient
	config             *Config
	gitserverLimiter   *rate.Limiter
	repoUpdaterLimiter *rate.Limiter
	operations         *operations
}

func NewIndexEnqueuer(
	dbStore DBStore,
	gitClient GitserverClient,
	repoUpdater RepoUpdaterClient,
	config *Config,
	observationContext *observation.Context,
) *IndexEnqueuer {
	return &IndexEnqueuer{
		dbStore:            dbStore,
		gitserverClient:    gitClient,
		repoUpdater:        repoUpdater,
		config:             config,
		gitserverLimiter:   rate.NewLimiter(config.MaximumRepositoriesInspectedPerSecond, 1),
		repoUpdaterLimiter: rate.NewLimiter(config.MaximumRepositoriesUpdatedPerSecond, 1),
		operations:         newOperations(observationContext),
	}
}

// InferIndexConfiguration looks at the repository contents at the lastest commit on the default branch of the given
// repository and determines an index configuration that is likely to succeed.
func (s *IndexEnqueuer) InferIndexConfiguration(ctx context.Context, repositoryID int) (_ *config.IndexConfiguration, err error) {
	ctx, trace, endObservation := s.operations.InferIndexConfiguration.WithAndLogger(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", repositoryID),
		},
	})
	defer endObservation(1, observation.Args{})

	commit, ok, err := s.gitserverClient.Head(ctx, repositoryID)
	if err != nil || !ok {
		return nil, errors.Wrap(err, "gitserver.Head")
	}
	trace.Log(log.String("commit", commit))

	indexJobs, err := s.inferIndexJobsFromRepositoryStructure(ctx, repositoryID, commit)
	if err != nil || len(indexJobs) == 0 {
		return nil, err
	}

	return &config.IndexConfiguration{
		IndexJobs: indexJobs,
	}, nil
}

// QueueIndexes enqueues a set of index jobs for the following repository and commit. If a non-empty
// configuration is given, it will be used to determine the set of jobs to enqueue. Otherwise, it will
// the configuration will be determined based on the regular index scheduling rules: first read any
// in-repo configuration (e.g., sourcegraph.yaml), then look for any existing in-database configuration,
// finally falling back to the automatically inferred connfiguration based on the repo contents at the
// target commit.
//
// If the force flag is false, then the presence of an upload or index record for this given repository and commit
// will cause this method to no-op. Note that this is NOT a guarantee that there will never be any duplicate records
// when the flag is false.
func (s *IndexEnqueuer) QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force bool) (_ []store.Index, err error) {
	ctx, trace, endObservation := s.operations.QueueIndex.WithAndLogger(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", repositoryID),
		},
	})
	defer endObservation(1, observation.Args{})

	commitID, err := s.gitserverClient.ResolveRevision(ctx, repositoryID, rev)
	if err != nil {
		return nil, errors.Wrap(err, "gitserver.ResolveRevision")
	}
	commit := string(commitID)
	trace.Log(log.String("commit", commit))

	return s.queueIndexForRepositoryAndCommit(ctx, repositoryID, commit, configuration, force, trace)
}

// QueueIndexesForPackage enqueues index jobs for a dependency of a recently-processed precise code
// intelligence index.
func (s *IndexEnqueuer) QueueIndexesForPackage(ctx context.Context, pkg precise.Package) (err error) {
	ctx, trace, endObservation := s.operations.QueueIndexForPackage.WithAndLogger(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.String("scheme", pkg.Scheme),
			log.String("name", pkg.Name),
			log.String("version", pkg.Version),
		},
	})
	defer endObservation(1, observation.Args{})

	repoName, revision, ok := InferRepositoryAndRevision(pkg)
	if !ok {
		return nil
	}
	trace.Log(log.String("repoName", repoName))
	trace.Log(log.String("revision", revision))

	if err := s.repoUpdaterLimiter.Wait(ctx); err != nil {
		return err
	}

	resp, err := s.repoUpdater.EnqueueRepoUpdate(ctx, api.RepoName(repoName))
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil
		}

		return errors.Wrap(err, "repoUpdater.EnqueueRepoUpdate")
	}

	commit, err := s.gitserverClient.ResolveRevision(ctx, int(resp.ID), revision)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil
		}

		return errors.Wrap(err, "gitserverClient.ResolveRevision")
	}

	_, err = s.queueIndexForRepositoryAndCommit(ctx, int(resp.ID), string(commit), "", false, trace)
	return err
}

// queueIndexForRepositoryAndCommit determines a set of index jobs to enqueue for the given repository and commit.
//
// If the force flag is false, then the presence of an upload or index record for this given repository and commit
// will cause this method to no-op. Note that this is NOT a guarantee that there will never be any duplicate records
// when the flag is false.
func (s *IndexEnqueuer) queueIndexForRepositoryAndCommit(ctx context.Context, repositoryID int, commit, configuration string, force bool, trace observation.TraceLogger) ([]store.Index, error) {
	if !force {
		isQueued, err := s.dbStore.IsQueued(ctx, repositoryID, commit)
		if err != nil {
			return nil, errors.Wrap(err, "dbstore.IsQueued")
		}
		if isQueued {
			return nil, nil
		}
	}

	indexes, err := s.getIndexRecords(ctx, repositoryID, commit, configuration)
	if err != nil {
		return nil, err
	}
	if len(indexes) == 0 {
		return nil, nil
	}
	trace.Log(log.Int("numIndexes", len(indexes)))

	return s.dbStore.InsertIndexes(ctx, indexes)
}

// inferIndexJobsFromRepositoryStructure collects the result of  InferIndexJobs over all registered recognizers.
func (s *IndexEnqueuer) inferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string) ([]config.IndexJob, error) {
	if err := s.gitserverLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	paths, err := s.gitserverClient.ListFiles(ctx, repositoryID, commit, inference.Patterns)
	if err != nil {
		return nil, errors.Wrap(err, "gitserver.ListFiles")
	}

	gitclient := newGitClient(s.gitserverClient, repositoryID, commit)

	var indexes []config.IndexJob
	for _, recognizer := range inference.Recognizers {
		recognizedPaths := []string{}
		pattern := inference.OrPattern(recognizer.Patterns())
		for _, path := range paths {
			if pattern.MatchString(path) {
				recognizedPaths = append(recognizedPaths, path)
			}
		}
		if len(recognizedPaths) > 0 {
			indexes = append(indexes, recognizer.InferIndexJobs(gitclient, recognizedPaths)...)
		}
	}

	if len(indexes) > s.config.MaximumIndexJobsPerInferredConfiguration {
		log15.Info("Too many inferred roots. Scheduling no index jobs for repository.", "repository_id", repositoryID)
		return nil, nil
	}

	return indexes, nil
}
