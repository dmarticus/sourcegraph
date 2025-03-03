package graphqlbackend

import (
	"context"
	"io/fs"
	"net/url"
	neturl "net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cloneurls"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/highlight"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var metricLabels = []string{"origin"}
var codeIntelRequests = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_lsif_requests",
	Help: "Counts LSIF requests.",
}, metricLabels)

// GitTreeEntryResolver resolves an entry in a Git tree in a repository. The entry can be any Git
// object type that is valid in a tree.
//
// Prefer using the constructor, NewGitTreeEntryResolver.
type GitTreeEntryResolver struct {
	db     database.DB
	commit *GitCommitResolver

	contentOnce sync.Once
	content     []byte
	contentErr  error

	// stat is this tree entry's file info. Its Name method must return the full path relative to
	// the root, not the basename.
	stat fs.FileInfo

	isRecursive   bool  // whether entries is populated recursively (otherwise just current level of hierarchy)
	isSingleChild *bool // whether this is the single entry in its parent. Only set by the (&GitTreeEntryResolver) entries.
}

func NewGitTreeEntryResolver(db database.DB, commit *GitCommitResolver, stat fs.FileInfo) *GitTreeEntryResolver {
	return &GitTreeEntryResolver{db: db, commit: commit, stat: stat}
}

func (r *GitTreeEntryResolver) Path() string { return r.stat.Name() }
func (r *GitTreeEntryResolver) Name() string { return path.Base(r.stat.Name()) }

func (r *GitTreeEntryResolver) ToGitTree() (*GitTreeEntryResolver, bool) { return r, true }
func (r *GitTreeEntryResolver) ToGitBlob() (*GitTreeEntryResolver, bool) { return r, true }

func (r *GitTreeEntryResolver) ToVirtualFile() (*virtualFileResolver, bool) { return nil, false }

func (r *GitTreeEntryResolver) ByteSize(ctx context.Context) (int32, error) {
	content, err := r.Content(ctx)
	if err != nil {
		return 0, err
	}
	return int32(len([]byte(content))), nil
}

func (r *GitTreeEntryResolver) Content(ctx context.Context) (string, error) {
	r.contentOnce.Do(func() {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		r.content, r.contentErr = git.ReadFile(
			ctx,
			r.commit.repoResolver.RepoName(),
			api.CommitID(r.commit.OID()),
			r.Path(),
			0,
			authz.DefaultSubRepoPermsChecker,
		)
	})

	return string(r.content), r.contentErr
}

func (r *GitTreeEntryResolver) RichHTML(ctx context.Context) (string, error) {
	content, err := r.Content(ctx)
	if err != nil {
		return "", err
	}
	return richHTML(content, path.Ext(r.Path()))
}

func (r *GitTreeEntryResolver) Binary(ctx context.Context) (bool, error) {
	content, err := r.Content(ctx)
	if err != nil {
		return false, err
	}
	return highlight.IsBinary([]byte(content)), nil
}

func (r *GitTreeEntryResolver) Highlight(ctx context.Context, args *HighlightArgs) (*highlightedFileResolver, error) {
	content, err := r.Content(ctx)
	if err != nil {
		return nil, err
	}
	return highlightContent(ctx, args, content, r.Path(), highlight.Metadata{
		RepoName: r.commit.repoResolver.Name(),
		Revision: string(r.commit.oid),
	})
}

func (r *GitTreeEntryResolver) Commit() *GitCommitResolver { return r.commit }

func (r *GitTreeEntryResolver) Repository() *RepositoryResolver { return r.commit.repoResolver }

func (r *GitTreeEntryResolver) IsRecursive() bool { return r.isRecursive }

func (r *GitTreeEntryResolver) URL(ctx context.Context) (string, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "treeentry.URL")
	defer span.Finish()

	if submodule := r.Submodule(); submodule != nil {
		span.SetTag("Submodule", "true")
		url := submodule.URL()
		if strings.HasPrefix(url, "../") {
			url = path.Join(r.Repository().Name(), url)
		}
		repoName, err := cloneURLToRepoName(ctx, r.db, url)
		if err != nil {
			log15.Error("Failed to resolve submodule repository name from clone URL", "cloneURL", submodule.URL(), "err", err)
			return "", nil
		}
		return "/" + repoName + "@" + submodule.Commit(), nil
	}
	url := r.commit.repoRevURL()
	return r.urlPath(url).String(), nil
}

func (r *GitTreeEntryResolver) CanonicalURL() string {
	url := r.commit.canonicalRepoRevURL()
	return r.urlPath(url).String()
}

func (r *GitTreeEntryResolver) urlPath(prefix *url.URL) *url.URL {
	// Dereference to copy to avoid mutating the input
	u := *prefix
	if r.IsRoot() {
		return &u
	}

	typ := "blob"
	if r.IsDirectory() {
		typ = "tree"
	}

	u.Path = path.Join(u.Path, "-", typ, r.Path())
	return &u
}

func (r *GitTreeEntryResolver) IsDirectory() bool { return r.stat.Mode().IsDir() }

func (r *GitTreeEntryResolver) ExternalURLs(ctx context.Context) ([]*externallink.Resolver, error) {
	repo, err := r.commit.repoResolver.repo(ctx)
	if err != nil {
		return nil, err
	}
	return externallink.FileOrDir(ctx, r.db, repo, r.commit.inputRevOrImmutableRev(), r.Path(), r.stat.Mode().IsDir())
}

func (r *GitTreeEntryResolver) RawZipArchiveURL() string {
	return globals.ExternalURL().ResolveReference(&neturl.URL{
		Path:     path.Join(r.Repository().URL(), "-/raw/", r.Path()),
		RawQuery: "format=zip",
	}).String()
}

func (r *GitTreeEntryResolver) Submodule() *gitSubmoduleResolver {
	if submoduleInfo, ok := r.stat.Sys().(git.Submodule); ok {
		return &gitSubmoduleResolver{submodule: submoduleInfo}
	}
	return nil
}

func cloneURLToRepoName(ctx context.Context, db database.DB, cloneURL string) (string, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "cloneURLToRepoName")
	defer span.Finish()

	repoName, err := cloneurls.ReposourceCloneURLToRepoName(ctx, db, cloneURL)
	if err != nil {
		return "", err
	}
	if repoName == "" {
		return "", errors.Errorf("no matching code host found for %s", cloneURL)
	}
	return string(repoName), nil
}

func CreateFileInfo(path string, isDir bool) fs.FileInfo {
	return fileInfo{path: path, isDir: isDir}
}

func (r *GitTreeEntryResolver) IsSingleChild(ctx context.Context, args *gitTreeEntryConnectionArgs) (bool, error) {
	if !r.IsDirectory() {
		return false, nil
	}
	if r.isSingleChild != nil {
		return *r.isSingleChild, nil
	}
	entries, err := git.ReadDir(
		ctx,
		authz.DefaultSubRepoPermsChecker,
		r.commit.repoResolver.RepoName(),
		api.CommitID(r.commit.OID()),
		path.Dir(r.Path()),
		false,
	)
	if err != nil {
		return false, err
	}
	return len(entries) == 1, nil
}

func (r *GitTreeEntryResolver) LSIF(ctx context.Context, args *struct{ ToolName *string }) (GitBlobLSIFDataResolver, error) {
	codeIntelRequests.WithLabelValues(trace.RequestOrigin(ctx)).Inc()

	var toolName string
	if args.ToolName != nil {
		toolName = *args.ToolName
	}

	repo, err := r.commit.repoResolver.repo(ctx)
	if err != nil {
		return nil, err
	}

	return EnterpriseResolvers.codeIntelResolver.GitBlobLSIFData(ctx, &GitBlobLSIFDataArgs{
		Repo:      repo,
		Commit:    api.CommitID(r.Commit().OID()),
		Path:      r.Path(),
		ExactPath: !r.stat.IsDir(),
		ToolName:  toolName,
	})
}

type fileInfo struct {
	path  string
	size  int64
	isDir bool
}

func (f fileInfo) Name() string { return f.path }
func (f fileInfo) Size() int64  { return f.size }
func (f fileInfo) IsDir() bool  { return f.isDir }
func (f fileInfo) Mode() os.FileMode {
	if f.IsDir() {
		return os.ModeDir
	}
	return 0
}
func (f fileInfo) ModTime() time.Time { return time.Now() }
func (f fileInfo) Sys() interface{}   { return interface{}(nil) }
