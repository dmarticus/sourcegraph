package cli

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/tmpfriend"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/throttled/throttled/v2/store/redigostore"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/updatecheck"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/bg"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli/loghandlers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/siteid"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/vfsutil"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/sentry"
	"github.com/sourcegraph/sourcegraph/internal/sysreq"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	traceFields    = env.Get("SRC_LOG_TRACE", "HTTP", "space separated list of trace logs to show. Options: all, HTTP, build, github")
	traceThreshold = env.Get("SRC_LOG_TRACE_THRESHOLD", "", "show traces that take longer than this")

	printLogo, _ = strconv.ParseBool(env.Get("LOGO", "false", "print Sourcegraph logo upon startup"))

	httpAddr         = env.Get("SRC_HTTP_ADDR", ":3080", "HTTP listen address for app and HTTP API")
	httpAddrInternal = envvar.HTTPAddrInternal

	nginxAddr = env.Get("SRC_NGINX_HTTP_ADDR", "", "HTTP listen address for nginx reverse proxy to SRC_HTTP_ADDR. Has preference over SRC_HTTP_ADDR for ExternalURL.")

	// dev browser browser extension ID. You can find this by going to chrome://extensions
	devExtension = "chrome-extension://bmfbcejdknlknpncfpeloejonjoledha"
	// production browser extension ID. This is found by viewing our extension in the chrome store.
	prodExtension = "chrome-extension://dgjhfomjieaadpoljlnidmbgkdffpack"
)

func init() {
	// If CACHE_DIR is specified, use that
	cacheDir := env.Get("CACHE_DIR", "/tmp", "directory to store cached archives.")
	vfsutil.ArchiveCacheDir = filepath.Join(cacheDir, "frontend-archive-cache")
}

// defaultExternalURL returns the default external URL of the application.
func defaultExternalURL(nginxAddr, httpAddr string) *url.URL {
	addr := nginxAddr
	if addr == "" {
		addr = httpAddr
	}

	var hostPort string
	if strings.HasPrefix(addr, ":") {
		// Prepend localhost if HTTP listen addr is just a port.
		hostPort = "127.0.0.1" + addr
	} else {
		hostPort = addr
	}

	return &url.URL{Scheme: "http", Host: hostPort}
}

// InitDB initializes and returns the global database connection and sets the
// version of the frontend in our versions table.
func InitDB() (*sql.DB, error) {
	var (
		sqlDB *sql.DB
		err   error
	)
	if os.Getenv("NEW_MIGRATIONS") == "" {
		// CURRENTLY DEPRECATING
		sqlDB, err = connections.NewFrontendDB("", "frontend", true, &observation.TestContext)
	} else {
		sqlDB, err = connections.EnsureNewFrontendDB("", "frontend", &observation.TestContext)
	}
	if err != nil {
		return nil, errors.Errorf("failed to connect to frontend database: %s", err)
	}

	if err := backend.UpdateServiceVersion(context.Background(), database.NewDB(sqlDB), "frontend", version.Version()); err != nil {
		return nil, err
	}

	return sqlDB, nil
}

// Main is the main entrypoint for the frontend server program.
func Main(enterpriseSetupHook func(db database.DB, c conftypes.UnifiedWatchable) enterprise.Services) error {
	ctx := context.Background()

	log.SetFlags(0)
	log.SetPrefix("")

	if err := profiler.Init(); err != nil {
		log.Fatalf("failed to initialize profiling: %v", err)
	}

	ready := make(chan struct{})
	go debugserver.NewServerRoutine(ready).Start()

	sqlDB, err := InitDB()
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	db := database.NewDB(sqlDB)

	if os.Getenv("SRC_DISABLE_OOBMIGRATION_VALIDATION") != "" {
		log15.Warn("Skipping out-of-band migrations check")
	} else {
		observationContext := &observation.Context{
			Logger:     log15.Root(),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}
		outOfBandMigrationRunner := oobmigration.NewRunnerWithDB(db, oobmigration.RefreshInterval, observationContext)

		if err := oobmigration.ValidateOutOfBandMigrationRunner(ctx, db, outOfBandMigrationRunner); err != nil {
			log.Fatalf("failed to validate out of band migrations: %v", err)
		}
	}

	// override site config first
	if err := overrideSiteConfig(ctx, db); err != nil {
		log.Fatalf("failed to apply site config overrides: %v", err)
	}
	globals.ConfigurationServerFrontendOnly = conf.InitConfigurationServerFrontendOnly(&configurationSource{db: db})
	conf.Init()
	conf.MustValidateDefaults()

	// now we can init the keyring, as it depends on site config
	if err := keyring.Init(ctx); err != nil {
		log.Fatalf("failed to initialize encryption keyring: %v", err)
	}

	if err := overrideGlobalSettings(ctx, db); err != nil {
		log.Fatalf("failed to override global settings: %v", err)
	}

	// now the keyring is configured it's safe to override the rest of the config
	// and that config can access the keyring
	if err := overrideExtSvcConfig(ctx, db); err != nil {
		log.Fatalf("failed to override external service config: %v", err)
	}

	// Filter trace logs
	d, _ := time.ParseDuration(traceThreshold)
	logging.Init(logging.Filter(loghandlers.Trace(strings.Fields(traceFields), d)))
	tracer.Init(conf.DefaultClient())
	sentry.Init(conf.DefaultClient())
	trace.Init()

	// Run enterprise setup hook
	enterprise := enterpriseSetupHook(db, conf.DefaultClient())

	authz.DefaultSubRepoPermsChecker, err = authz.NewSubRepoPermsClient(database.SubRepoPerms(db))
	if err != nil {
		log.Fatalf("Failed to create sub-repo client: %v", err)
	}
	ui.InitRouter(db, enterprise.CodeIntelResolver)

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "help", "-h", "--help":
			log.Printf("Version: %s", version.Version())
			log.Print()

			env.PrintHelp()

			log.Print()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			for _, st := range sysreq.Check(ctx, skippedSysReqs()) {
				log.Printf("%s:", st.Name)
				if st.OK() {
					log.Print("\tOK")
					continue
				}
				if st.Skipped {
					log.Print("\tSkipped")
					continue
				}
				if st.Problem != "" {
					log.Print("\t" + st.Problem)
				}
				if st.Err != nil {
					log.Printf("\tError: %s", st.Err)
				}
				if st.Fix != "" {
					log.Printf("\tPossible fix: %s", st.Fix)
				}
			}

			return nil
		}
	}

	printConfigValidation()

	cleanup := tmpfriend.SetupOrNOOP()
	defer cleanup()

	// Don't proceed if system requirements are missing, to avoid
	// presenting users with a half-working experience.
	if err := checkSysReqs(context.Background(), os.Stderr); err != nil {
		return err
	}

	siteid.Init(db)

	globals.WatchExternalURL(defaultExternalURL(nginxAddr, httpAddr))
	globals.WatchPermissionsUserMapping()

	goroutine.Go(func() { bg.CheckRedisCacheEvictionPolicy() })
	goroutine.Go(func() { bg.DeleteOldCacheDataInRedis() })
	goroutine.Go(func() { bg.DeleteOldEventLogsInPostgres(context.Background(), db) })
	goroutine.Go(func() { bg.DeleteOldSecurityEventLogsInPostgres(context.Background(), db) })
	goroutine.Go(func() { updatecheck.Start(db) })

	schema, err := graphqlbackend.NewSchema(db,
		enterprise.BatchChangesResolver,
		enterprise.CodeIntelResolver,
		enterprise.InsightsResolver,
		enterprise.AuthzResolver,
		enterprise.CodeMonitorsResolver,
		enterprise.LicenseResolver,
		enterprise.DotcomResolver,
		enterprise.SearchContextsResolver,
		enterprise.OrgRepositoryResolver,
		enterprise.NotebooksResolver,
		enterprise.ComputeResolver,
	)
	if err != nil {
		return err
	}

	rateLimitWatcher, err := makeRateLimitWatcher()
	if err != nil {
		return err
	}

	server, err := makeExternalAPI(db, schema, enterprise, rateLimitWatcher)
	if err != nil {
		return err
	}

	internalAPI, err := makeInternalAPI(schema, db, enterprise, rateLimitWatcher)
	if err != nil {
		return err
	}

	routines := []goroutine.BackgroundRoutine{server}
	if internalAPI != nil {
		routines = append(routines, internalAPI)
	}

	if printLogo {
		fmt.Println(" ")
		fmt.Println(logoColor)
		fmt.Println(" ")
	}
	fmt.Printf("✱ Sourcegraph is ready at: %s\n", globals.ExternalURL())
	close(ready)

	goroutine.MonitorBackgroundRoutines(context.Background(), routines...)
	return nil
}

func makeExternalAPI(db database.DB, schema *graphql.Schema, enterprise enterprise.Services, rateLimiter graphqlbackend.LimitWatcher) (goroutine.BackgroundRoutine, error) {
	listener, err := httpserver.NewListener(httpAddr)
	if err != nil {
		return nil, err
	}

	// Create the external HTTP handler.
	externalHandler, err := newExternalHTTPHandler(
		db,
		schema,
		enterprise.GitHubWebhook,
		enterprise.GitLabWebhook,
		enterprise.BitbucketServerWebhook,
		enterprise.NewCodeIntelUploadHandler,
		enterprise.NewExecutorProxyHandler,
		enterprise.NewGitHubAppCloudSetupHandler,
		enterprise.NewComputeStreamHandler,
		rateLimiter,
	)
	if err != nil {
		return nil, err
	}
	httpServer := &http.Server{
		Handler:      externalHandler,
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
	}

	server := httpserver.New(listener, httpServer, makeServerOptions()...)
	log15.Debug("HTTP running", "on", httpAddr)
	return server, nil
}

func makeInternalAPI(schema *graphql.Schema, db database.DB, enterprise enterprise.Services, rateLimiter graphqlbackend.LimitWatcher) (goroutine.BackgroundRoutine, error) {
	if httpAddrInternal == "" {
		return nil, nil
	}

	listener, err := httpserver.NewListener(httpAddrInternal)
	if err != nil {
		return nil, err
	}

	// The internal HTTP handler does not include the auth handlers.
	internalHandler := newInternalHTTPHandler(
		schema,
		db,
		enterprise.NewCodeIntelUploadHandler,
		rateLimiter,
	)
	httpServer := &http.Server{
		Handler:     internalHandler,
		ReadTimeout: 75 * time.Second,
		// Higher since for internal RPCs which can have large responses
		// (eg git archive). Should match the timeout used for git archive
		// in gitserver.
		WriteTimeout: time.Hour,
	}

	server := httpserver.New(listener, httpServer, makeServerOptions()...)
	log15.Debug("HTTP (internal) running", "on", httpAddrInternal)
	return server, nil
}

func makeServerOptions() (options []httpserver.ServerOptions) {
	if deploy.Type() == deploy.Kubernetes {
		// On kubernetes, we want to wait an additional 5 seconds after we receive a
		// shutdown request to give some additional time for the endpoint changes
		// to propagate to services talking to this server like the LB or ingress
		// controller. We only do this in frontend and not on all services, because
		// frontend is the only publicly exposed service where we don't control
		// retries on connection failures (see httpcli.InternalClient).
		options = append(options, httpserver.WithPreShutdownPause(time.Second*5))
	}

	return options
}

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, o := range allowedOrigins {
		if o == "*" || o == origin {
			return true
		}
	}
	return false
}

func makeRateLimitWatcher() (*graphqlbackend.BasicLimitWatcher, error) {
	ratelimitStore, err := redigostore.New(redispool.Cache, "gql:rl:", 0)
	if err != nil {
		return nil, err
	}
	return graphqlbackend.NewBasicLimitWatcher(ratelimitStore), nil
}
