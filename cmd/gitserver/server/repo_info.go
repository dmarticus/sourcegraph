package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *Server) repoInfo(ctx context.Context, repo api.RepoName) (*protocol.RepoInfo, error) {
	dir := s.dir(repo)
	resp := protocol.RepoInfo{
		Cloned: repoCloned(dir),
	}
	if resp.Cloned {
		// TODO(keegancsmith,tsenart) the only user of this information is the site admin
		// settings page for a repo. That page should just ask the DB for the remote URL.
		//
		// We need an internal actor here since we query the repo table. We are trusting
		// the auth checks we already have in place that do not allow a site admin to
		// view private repos they don't own.
		remoteURL, err := s.getRemoteURL(actor.WithInternalActor(ctx), repo)
		if err != nil {
			return nil, err
		}
		resp.URL = remoteURL.String()
	}
	{
		resp.CloneProgress, resp.CloneInProgress = s.locker.Status(dir)
		if isAlwaysCloningTest(repo) {
			resp.CloneInProgress = true
			resp.CloneProgress = "This will never finish cloning"
		}
	}
	if resp.Cloned {
		if mtime, err := repoLastFetched(dir); err != nil {
			log15.Warn("error computing last-fetched date", "repo", repo, "err", err)
		} else {
			resp.LastFetched = &mtime
		}

		if cloneTime, err := getRecloneTime(dir); err != nil {
			log15.Warn("error getting re-clone time", "repo", repo, "err", err)
		} else {
			resp.CloneTime = &cloneTime
		}

		if lastChanged, err := repoLastChanged(dir); err != nil {
			log15.Warn("error getting last changed", "repo", repo, "err", err)
		} else {
			resp.LastChanged = &lastChanged
		}
	}
	return &resp, nil
}

func (s *Server) repoCloneProgress(repo api.RepoName) (*protocol.RepoCloneProgress, error) {
	dir := s.dir(repo)
	resp := protocol.RepoCloneProgress{
		Cloned: repoCloned(dir),
	}
	resp.CloneProgress, resp.CloneInProgress = s.locker.Status(dir)
	if isAlwaysCloningTest(repo) {
		resp.CloneInProgress = true
		resp.CloneProgress = "This will never finish cloning"
	}
	return &resp, nil
}

func (s *Server) handleRepoInfo(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := protocol.RepoInfoResponse{
		Results: make(map[api.RepoName]*protocol.RepoInfo, len(req.Repos)),
	}
	for _, repoName := range req.Repos {
		result, err := s.repoInfo(r.Context(), repoName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.Results[repoName] = result
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleReposStats(w http.ResponseWriter, r *http.Request) {
	b, err := os.ReadFile(filepath.Join(s.ReposDir, reposStatsName))
	if errors.Is(err, os.ErrNotExist) {
		// When a gitserver is new this file might not have been computed
		// yet. Clients are expected to handle this case by noticing UpdatedAt
		// is not set.
		b = []byte("{}")
	} else if err != nil {
		http.Error(w, fmt.Sprintf("failed to read %s: %v", reposStatsName, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(b)
}

func (s *Server) handleRepoCloneProgress(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoCloneProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := protocol.RepoCloneProgressResponse{
		Results: make(map[api.RepoName]*protocol.RepoCloneProgress, len(req.Repos)),
	}
	for _, repoName := range req.Repos {
		result, err := s.repoCloneProgress(repoName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.Results[repoName] = result
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleRepoDelete(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.deleteRepo(req.Repo); err != nil {
		log15.Error("failed to delete repository", "repo", req.Repo, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log15.Info("deleted repository", "repo", req.Repo)
}

func (s *Server) deleteRepo(repo api.RepoName) error {
	return s.removeRepoDirectory(s.dir(repo))
}
