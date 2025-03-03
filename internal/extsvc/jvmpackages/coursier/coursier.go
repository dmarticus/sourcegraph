package coursier

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var CoursierBinary = "coursier"

var (
	coursierCacheDir   string
	observationContext *observation.Context
	operations         *Operations
	invocTimeout, _    = time.ParseDuration(env.Get("SRC_COURSIER_TIMEOUT", "2m", "Time limit per Coursier invocation, which is used to resolve JVM/Java dependencies."))
)

func init() {
	observationContext = &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}
	operations = NewOperations(observationContext)

	// Should only be set for gitserver for persistence, repo-updater will use ephemeral storage.
	// repo-updater only performs existence checks which doesnt involve downloading any JARs (except for JDK),
	// only POM files which are much lighter.
	if reposDir := os.Getenv("SRC_REPOS_DIR"); reposDir != "" {
		coursierCacheDir = filepath.Join(reposDir, "coursier")
		if err := os.MkdirAll(coursierCacheDir, os.ModePerm); err != nil {
			log.Fatalf("failed to create coursier cache dir in %s: %s", coursierCacheDir, err)
		}
	}
}

func FetchSources(ctx context.Context, config *schema.JVMPackagesConnection, dependency reposource.MavenDependency) (sourceCodeJarPath string, err error) {
	ctx, endObservation := operations.fetchSources.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("dependency", dependency.PackageManagerSyntax()),
	}})
	defer endObservation(1, observation.Args{})

	if dependency.IsJDK() {
		output, err := runCoursierCommand(
			ctx,
			config,
			"java-home", "--jvm",
			dependency.Version,
		)
		if err != nil {
			return "", err
		}
		for _, outputPath := range output {
			for _, srcPath := range []string{
				path.Join(outputPath, "src.zip"),
				path.Join(outputPath, "lib", "src.zip"),
			} {
				stat, err := os.Stat(srcPath)
				if !os.IsNotExist(err) && stat.Mode().IsRegular() {
					return srcPath, nil
				}
			}
		}
		return "", errors.Errorf("failed to find src.zip for JVM dependency %s", dependency)
	}
	paths, err := runCoursierCommand(
		ctx,
		config,
		// NOTE: make sure to update the method `coursierScript` in
		// vcs_syncer_jvm_packages_test.go if you change the arguments
		// here. The test case assumes that the "--classifier sources"
		// arguments appears at a specific index.
		"fetch",
		"--quiet", "--quiet",
		"--intransitive", dependency.PackageManagerSyntax(),
		"--classifier", "sources",
	)
	if err != nil {
		return "", err
	}
	if len(paths) == 0 || (len(paths) == 1 && paths[0] == "") {
		return "", ErrNoSources{Dependency: dependency}
	}
	if len(paths) > 1 {
		return "", errors.Errorf("expected single JAR path but found multiple: %v", paths)
	}
	return paths[0], nil
}

// ErrNoSources indicates that a dependency has no sources
type ErrNoSources struct {
	Dependency reposource.MavenDependency
}

func (e ErrNoSources) Error() string {
	return fmt.Sprintf("no sources for dependency %s", e.Dependency)
}

func FetchByteCode(ctx context.Context, config *schema.JVMPackagesConnection, dependency reposource.MavenDependency) (byteCodeJarPath string, err error) {
	ctx, endObservation := operations.fetchByteCode.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	paths, err := runCoursierCommand(
		ctx,
		config,
		// NOTE: make sure to update the method `coursierScript` in
		// vcs_syncer_jvm_packages_test.go if you change the arguments
		// here. The test case assumes that the "--classifier sources"
		// arguments appears at a specific index.
		"fetch",
		"--quiet", "--quiet",
		"--intransitive", dependency.PackageManagerSyntax(),
	)
	if err != nil {
		return "", err
	}
	if len(paths) == 0 || (paths[0] == "") {
		return "", errors.Errorf("no bytecode jar for dependency %s", dependency)
	}
	if len(paths) > 1 {
		return "", errors.Errorf("expected single JAR path but found multiple: %v", paths)
	}
	return paths[0], nil
}

func Exists(ctx context.Context, config *schema.JVMPackagesConnection, dependency reposource.MavenDependency) (exists bool, err error) {
	ctx, endObservation := operations.exists.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("dependency", dependency.PackageManagerSyntax()),
	}})
	defer endObservation(1, observation.Args{})

	if dependency.IsJDK() {
		_, err = FetchSources(ctx, config, dependency)
		return err == nil, err
	}
	_, err = runCoursierCommand(
		ctx,
		config,
		"resolve",
		"--quiet", "--quiet",
		"--intransitive", dependency.PackageManagerSyntax(),
	)
	return err == nil, err
}

func runCoursierCommand(ctx context.Context, config *schema.JVMPackagesConnection, args ...string) (stdoutLines []string, err error) {
	ctx, cancel := context.WithTimeout(ctx, invocTimeout)
	defer cancel()

	ctx, trace, endObservation := operations.runCommand.WithAndLogger(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("repositories", strings.Join(config.Maven.Repositories, "|")),
		otlog.String("args", strings.Join(args, ", ")),
	}})
	defer endObservation(1, observation.Args{})

	cmd := exec.CommandContext(ctx, CoursierBinary, args...)
	if config.Maven.Credentials != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("COURSIER_CREDENTIALS=%v", config.Maven.Credentials))
	}
	if len(config.Maven.Repositories) > 0 {
		cmd.Env = append(
			cmd.Env,
			fmt.Sprintf("COURSIER_REPOSITORIES=%v", strings.Join(config.Maven.Repositories, "|")),
		)
	}
	if coursierCacheDir != "" {
		cmd.Env = append(cmd.Env, "COURSIER_CACHE="+coursierCacheDir)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "coursier command %q failed with stderr %q and stdout %q", cmd, stderr, &stdout)
	}
	trace.Log(otlog.String("stdout", stdout.String()), otlog.String("stderr", stderr.String()))

	if stdout.String() == "" {
		return []string{}, nil
	}

	return strings.Split(strings.TrimSpace(stdout.String()), "\n"), nil
}
