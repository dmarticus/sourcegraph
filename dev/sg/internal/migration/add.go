package migration

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const newMetadataFileTemplate = `name: %s
parents: [%s]
`

const newUpMigrationFileTemplate = `BEGIN;

-- Perform migration here.
--
-- See /migrations/README.md. Highlights:
--  * Make migrations idempotent (use IF EXISTS)
--  * Make migrations backwards-compatible (old readers/writers must continue to work)
--  * Wrap your changes in a transaction
--  * If you are using CREATE INDEX CONCURRENTLY, then make sure that only one statement
--    is defined per file, and that each such statement is NOT wrapped in a transaction.
--    Each such migration must also declare "createIndexConcurrently: true" in their
--    associated metadata.yaml file.

COMMIT;
`

const newDownMigrationFileTemplate = `BEGIN;

-- Undo the changes made in the up migration

COMMIT;
`

// Add creates a new directory with stub migration files in the given schema and returns the
// names of the newly created files. If there was an error, the filesystem is rolled-back.
func Add(database db.Database, migrationName string) error {
	return add(database, migrationName, newUpMigrationFileTemplate, newDownMigrationFileTemplate)
}

func add(database db.Database, migrationName, upMigrationFileTemplate, downMigrationFileTemplate string) error {
	definitions, err := readDefinitions(database)
	if err != nil {
		return err
	}

	leaves := definitions.Leaves()
	parents := make([]int, 0, len(leaves))
	for _, leaf := range leaves {
		parents = append(parents, leaf.ID)
	}

	files, err := makeMigrationFilenames(database, int(time.Now().UTC().Unix()))
	if err != nil {
		return err
	}

	contents := map[string]string{
		files.UpFile:       upMigrationFileTemplate,
		files.DownFile:     downMigrationFileTemplate,
		files.MetadataFile: fmt.Sprintf(newMetadataFileTemplate, migrationName, strings.Join(intsToStrings(parents), ", ")),
	}
	if err := writeMigrationFiles(contents); err != nil {
		return err
	}

	block := stdout.Out.Block(output.Linef("", output.StyleBold, "Migration files created"))
	block.Writef("Up query file: %s", files.UpFile)
	block.Writef("Down query file: %s", files.DownFile)
	block.Writef("Metadata file: %s", files.MetadataFile)
	block.Close()

	return nil
}

func intsToStrings(ints []int) []string {
	strs := make([]string, 0, len(ints))
	for _, value := range ints {
		strs = append(strs, strconv.Itoa(value))
	}

	return strs
}
