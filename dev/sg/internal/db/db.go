package db

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

func getFSForPath(path string) func() (fs.FS, error) {
	return func() (fs.FS, error) {
		repoRoot, err := root.RepositoryRoot()
		if err != nil {
			return nil, err
		}

		return os.DirFS(filepath.Join(repoRoot, "migrations", path)), nil
	}
}

type Database struct {
	// Name of database, used to convert from arguments to Database
	Name string

	// Table in database for storing information about migrations.
	MigrationsTable string

	// Additional data tables for database
	DataTables []string

	// Used for retrieving the directory where migrations live
	FS func() (fs.FS, error)
}

var (
	frontendDatabase = Database{
		Name:            "frontend",
		MigrationsTable: "schema_migrations",
		DataTables:      []string{"out_of_band_migrations"},
		FS:              getFSForPath("frontend"),
	}

	codeIntelDatabase = Database{
		Name:            "codeintel",
		MigrationsTable: "codeintel_schema_migrations",
		DataTables:      nil,
		FS:              getFSForPath("codeintel"),
	}

	codeInsightsDatabase = Database{
		Name:            "codeinsights",
		MigrationsTable: "codeinsights_schema_migrations",
		DataTables:      nil,
		FS:              getFSForPath("codeinsights"),
	}

	databases = []Database{
		frontendDatabase,
		codeIntelDatabase,
		codeInsightsDatabase,
	}

	DefaultDatabase = databases[0]
)

func Databases() []Database {
	c := make([]Database, len(databases))
	copy(c, databases)
	return c
}

func DatabaseNames() []string {
	databaseNames := make([]string, 0, len(databases))
	for _, database := range databases {
		databaseNames = append(databaseNames, database.Name)
	}
	sort.Strings(databaseNames)

	return databaseNames
}

func DatabaseByName(name string) (Database, bool) {
	for _, database := range databases {
		if database.Name == name {
			return database, true
		}
	}

	return Database{}, false
}
