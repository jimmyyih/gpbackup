package backup

import (
	"fmt"
	"path"

	"github.com/greenplum-db/gp-common-go-libs/dbconn"
	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gp-common-go-libs/iohelper"
	"github.com/greenplum-db/gpbackup/history"
	"github.com/greenplum-db/gpbackup/options"
	"github.com/greenplum-db/gpbackup/toc"
	"github.com/greenplum-db/gpbackup/utils"
	"github.com/pkg/errors"
)

func FilterTablesForIncremental(lastBackupTOC, currentTOC *toc.TOC, tables []Table) []Table {
	var filteredTables []Table
	for _, table := range tables {
		currentAOEntry, isAOTable := currentTOC.IncrementalMetadata.AO[table.FQN()]
		if !isAOTable {
			filteredTables = append(filteredTables, table)
			continue
		}
		previousAOEntry := lastBackupTOC.IncrementalMetadata.AO[table.FQN()]

		if previousAOEntry.Modcount != currentAOEntry.Modcount || previousAOEntry.LastDDLTimestamp != currentAOEntry.LastDDLTimestamp {
			filteredTables = append(filteredTables, table)
		}
	}

	return filteredTables
}

func GetTargetBackupTimestamp() string {
	targetTimestamp := ""
	if fromTimestamp := MustGetFlagString(options.FROM_TIMESTAMP); fromTimestamp != "" {
		validateFromTimestamp(fromTimestamp)
		targetTimestamp = fromTimestamp
	} else {
		targetTimestamp = GetLatestMatchingBackupTimestamp()
	}
	return targetTimestamp
}

func GetLatestMatchingBackupTimestamp() string {
	latestTimestamp := ""
	var contents *history.History
	var latestMatchingBackupHistoryEntry *history.BackupConfig
	var err error
	if iohelper.FileExistsAndIsReadable(globalFPInfo.GetBackupHistoryFilePath()) {
		contents, err = history.NewHistory(globalFPInfo.GetBackupHistoryFilePath())
		gplog.FatalOnError(err)
		latestMatchingBackupHistoryEntry = GetLatestMatchingBackupConfig(contents, &backupReport.BackupConfig)
	}

	if latestMatchingBackupHistoryEntry == nil {
		gplog.FatalOnError(errors.Errorf("There was no matching previous backup found with the flags provided. " +
			"Please take a full backup."))
	} else {
		latestTimestamp = latestMatchingBackupHistoryEntry.Timestamp
	}

	return latestTimestamp
}

func GetLatestMatchingBackupConfig(history *history.History, currentBackupConfig *history.BackupConfig) *history.BackupConfig {
	for _, backupConfig := range history.BackupConfigs {
		if matchesIncrementalFlags(&backupConfig, currentBackupConfig) {
			return &backupConfig
		}
	}

	return nil
}

func matchesIncrementalFlags(backupConfig *history.BackupConfig, currentBackupConfig *history.BackupConfig) bool {
	_, pluginBinaryName := path.Split(backupConfig.Plugin)
	return backupConfig.BackupDir == MustGetFlagString(options.BACKUP_DIR) &&
		backupConfig.DatabaseName == currentBackupConfig.DatabaseName &&
		backupConfig.LeafPartitionData == MustGetFlagBool(options.LEAF_PARTITION_DATA) &&
		pluginBinaryName == currentBackupConfig.Plugin &&
		backupConfig.SingleDataFile == MustGetFlagBool(options.SINGLE_DATA_FILE) &&
		backupConfig.Compressed == currentBackupConfig.Compressed &&
		// Expanding of the include list happens before this now so we must compare again current backup config
		utils.NewIncludeSet(backupConfig.IncludeRelations).Equals(utils.NewIncludeSet(currentBackupConfig.IncludeRelations)) &&
		utils.NewIncludeSet(backupConfig.IncludeSchemas).Equals(utils.NewIncludeSet(MustGetFlagStringArray(options.INCLUDE_SCHEMA))) &&
		utils.NewIncludeSet(backupConfig.ExcludeRelations).Equals(utils.NewIncludeSet(MustGetFlagStringArray(options.EXCLUDE_RELATION))) &&
		utils.NewIncludeSet(backupConfig.ExcludeSchemas).Equals(utils.NewIncludeSet(MustGetFlagStringArray(options.EXCLUDE_SCHEMA)))
}

func PopulateRestorePlan(changedTables []Table,
	restorePlan []history.RestorePlanEntry, allTables []Table) []history.RestorePlanEntry {
	currBackupRestorePlanEntry := history.RestorePlanEntry{
		Timestamp: globalFPInfo.Timestamp,
		TableFQNs: make([]string, 0, len(changedTables)),
	}

	for _, changedTable := range changedTables {
		changedTableFQN := changedTable.FQN()
		currBackupRestorePlanEntry.TableFQNs = append(currBackupRestorePlanEntry.TableFQNs, changedTableFQN)
	}

	changedTableFQNs := make(map[string]bool)
	for _, changedTable := range changedTables {
		changedTableFQN := changedTable.FQN()
		changedTableFQNs[changedTableFQN] = true
	}

	allTableFQNs := make(map[string]bool)
	for _, table := range allTables {
		tableFQN := table.FQN()
		allTableFQNs[tableFQN] = true
	}

	// Removing filtered table FQNs for the current backup from entries with previous timestamps
	for i, restorePlanEntry := range restorePlan {
		tableFQNs := make([]string, 0)
		for _, tableFQN := range restorePlanEntry.TableFQNs {
			if !changedTableFQNs[tableFQN] && allTableFQNs[tableFQN] {
				tableFQNs = append(tableFQNs, tableFQN)
			}
		}
		restorePlan[i].TableFQNs = tableFQNs
	}
	restorePlan = append(restorePlan, currBackupRestorePlanEntry)

	return restorePlan
}

// Incremental partitions
type partition struct {
	PartitionOid uint32 `db:"paroid"`
	ClassOid     uint32 `db:"parchildrelid"`
	// partitionName         string `db:"parname"`
	// isDefault             bool   `db:"parisdefault"`
	// rank                  int    `db:"parruleord"`
	// isRangeStartInclusive bool   `db:"parrangestartincl"`
	// isRangeEndInclusive   bool   `db:"parrangeendincl"`
	// rangeStart            string `db:"parrangestart"`
	// rangeEnd              string `db:"parrangeend"`
}

func GetIncrementalPartitionRestoreMetadata(conn *dbconn.DBConn) (map[uint32]partition, error) {
	query := fmt.Sprintf(`
	SELECT p.paroid FROM pg_partition_rule p;`)

	results := make([]partition, 0)
	err := conn.Select(&results, query)
	if err != nil {
		return nil, err
	}

	partitionRules := make(map[uint32]partition, 0)
	fmt.Println("==================================")
	fmt.Println(len(results))
	fmt.Println()
	for _, v := range results {
		fmt.Println(v)
		// partitionRules[v.classOid] = v
	}

	return partitionRules, nil
}
