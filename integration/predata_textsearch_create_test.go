package integration

import (
	"github.com/greenplum-db/gp-common-go-libs/structmatcher"
	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpbackup/backup"
	"github.com/greenplum-db/gpbackup/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("backup integration create statement tests", func() {
	BeforeEach(func() {
		testutils.SkipIfBefore5(connectionPool)
		toc, backupfile = testutils.InitializeTestTOC(buffer, "predata")
	})
	Describe("PrintCreateTextSearchParserStatements", func() {
		parser := backup.TextSearchParser{Oid: 0, Schema: "public", Name: "testparser", StartFunc: "prsd_start", TokenFunc: "prsd_nexttoken", EndFunc: "prsd_end", LexTypesFunc: "prsd_lextype", HeadlineFunc: "prsd_headline"}
		It("creates a basic text search parser", func() {
			backup.PrintCreateTextSearchParserStatement(backupfile, toc, parser, backup.ObjectMetadata{})

			testhelper.AssertQueryRuns(connectionPool, buffer.String())
			defer testhelper.AssertQueryRuns(connectionPool, "DROP TEXT SEARCH PARSER public.testparser")

			resultParsers := backup.GetTextSearchParsers(connectionPool)

			Expect(resultParsers).To(HaveLen(1))
			structmatcher.ExpectStructsToMatchExcluding(&parser, &resultParsers[0], "Oid")
		})
		It("creates a basic text search parser with a comment", func() {
			parserMetadata := testutils.DefaultMetadata("TEXT SEARCH PARSER", false, false, true)

			backup.PrintCreateTextSearchParserStatement(backupfile, toc, parser, parserMetadata)

			testhelper.AssertQueryRuns(connectionPool, buffer.String())
			defer testhelper.AssertQueryRuns(connectionPool, "DROP TEXT SEARCH PARSER public.testparser")

			resultParsers := backup.GetTextSearchParsers(connectionPool)
			resultMetadataMap := backup.GetCommentsForObjectType(connectionPool, backup.TYPE_TSPARSER)

			Expect(resultParsers).To(HaveLen(1))
			structmatcher.ExpectStructsToMatchExcluding(&parser, &resultParsers[0], "Oid")
			resultMetadata := resultMetadataMap[resultParsers[0].GetUniqueID()]
			structmatcher.ExpectStructsToMatch(&parserMetadata, &resultMetadata)
		})
	})
	Describe("PrintCreateTextSearchTemplateStatement", func() {
		template := backup.TextSearchTemplate{Oid: 1, Schema: "public", Name: "testtemplate", InitFunc: "dsimple_init", LexizeFunc: "dsimple_lexize"}
		It("creates a basic text search template", func() {
			backup.PrintCreateTextSearchTemplateStatement(backupfile, toc, template, backup.ObjectMetadata{})

			testhelper.AssertQueryRuns(connectionPool, buffer.String())
			defer testhelper.AssertQueryRuns(connectionPool, "DROP TEXT SEARCH TEMPLATE public.testtemplate")

			resultTemplates := backup.GetTextSearchTemplates(connectionPool)

			Expect(resultTemplates).To(HaveLen(1))
			structmatcher.ExpectStructsToMatchExcluding(&template, &resultTemplates[0], "Oid")
		})
		It("creates a basic text search template with a comment", func() {
			templateMetadata := testutils.DefaultMetadata("TEXT SEARCH TEMPLATE", false, false, true)

			backup.PrintCreateTextSearchTemplateStatement(backupfile, toc, template, templateMetadata)

			testhelper.AssertQueryRuns(connectionPool, buffer.String())
			defer testhelper.AssertQueryRuns(connectionPool, "DROP TEXT SEARCH TEMPLATE public.testtemplate")

			resultTemplates := backup.GetTextSearchTemplates(connectionPool)
			resultMetadataMap := backup.GetCommentsForObjectType(connectionPool, backup.TYPE_TSTEMPLATE)

			Expect(resultTemplates).To(HaveLen(1))
			uniqueID := testutils.UniqueIDFromObjectName(connectionPool, "public", "testtemplate", backup.TYPE_TSTEMPLATE)
			resultMetadata := resultMetadataMap[uniqueID]
			structmatcher.ExpectStructsToMatchExcluding(&template, &resultTemplates[0], "Oid")
			structmatcher.ExpectStructsToMatch(&templateMetadata, &resultMetadata)
		})
	})
	Describe("PrintCreateTextSearchDictionaryStatement", func() {
		dictionary := backup.TextSearchDictionary{Oid: 1, Schema: "public", Name: "testdictionary", Template: "pg_catalog.snowball", InitOption: "language = 'russian', stopwords = 'russian'"}
		It("creates a basic text search dictionary", func() {

			backup.PrintCreateTextSearchDictionaryStatement(backupfile, toc, dictionary, backup.ObjectMetadata{})

			testhelper.AssertQueryRuns(connectionPool, buffer.String())
			defer testhelper.AssertQueryRuns(connectionPool, "DROP TEXT SEARCH DICTIONARY public.testdictionary")

			resultDictionaries := backup.GetTextSearchDictionaries(connectionPool)

			Expect(resultDictionaries).To(HaveLen(1))
			structmatcher.ExpectStructsToMatchExcluding(&dictionary, &resultDictionaries[0], "Oid")
		})
		It("creates a basic text search dictionary with a comment and owner", func() {
			dictionaryMetadata := testutils.DefaultMetadata("TEXT SEARCH DICTIONARY", false, true, true)

			backup.PrintCreateTextSearchDictionaryStatement(backupfile, toc, dictionary, dictionaryMetadata)

			testhelper.AssertQueryRuns(connectionPool, buffer.String())
			defer testhelper.AssertQueryRuns(connectionPool, "DROP TEXT SEARCH DICTIONARY public.testdictionary")

			resultDictionaries := backup.GetTextSearchDictionaries(connectionPool)
			resultMetadataMap := backup.GetMetadataForObjectType(connectionPool, backup.TYPE_TSDICTIONARY)

			Expect(resultDictionaries).To(HaveLen(1))
			uniqueID := testutils.UniqueIDFromObjectName(connectionPool, "public", "testdictionary", backup.TYPE_TSDICTIONARY)
			resultMetadata := resultMetadataMap[uniqueID]
			structmatcher.ExpectStructsToMatchExcluding(&dictionary, &resultDictionaries[0], "Oid")
			structmatcher.ExpectStructsToMatch(&dictionaryMetadata, &resultMetadata)
		})
	})
	Describe("PrintCreateTextSearchConfigurationStatement", func() {
		configuration := backup.TextSearchConfiguration{Oid: 1, Schema: "public", Name: "testconfiguration", Parser: `pg_catalog."default"`, TokenToDicts: map[string][]string{}}
		It("creates a basic text search configuration", func() {
			backup.PrintCreateTextSearchConfigurationStatement(backupfile, toc, configuration, backup.ObjectMetadata{})

			testhelper.AssertQueryRuns(connectionPool, buffer.String())
			defer testhelper.AssertQueryRuns(connectionPool, "DROP TEXT SEARCH CONFIGURATION public.testconfiguration")

			resultConfigurations := backup.GetTextSearchConfigurations(connectionPool)

			Expect(resultConfigurations).To(HaveLen(1))
			structmatcher.ExpectStructsToMatchExcluding(&configuration, &resultConfigurations[0], "Oid")
		})
		It("creates a basic text search configuration with a comment and owner", func() {
			configurationMetadata := testutils.DefaultMetadata("TEXT SEARCH CONFIGURATION", false, true, true)

			backup.PrintCreateTextSearchConfigurationStatement(backupfile, toc, configuration, configurationMetadata)

			testhelper.AssertQueryRuns(connectionPool, buffer.String())
			defer testhelper.AssertQueryRuns(connectionPool, "DROP TEXT SEARCH CONFIGURATION public.testconfiguration")

			resultConfigurations := backup.GetTextSearchConfigurations(connectionPool)
			resultMetadataMap := backup.GetMetadataForObjectType(connectionPool, backup.TYPE_TSCONFIGURATION)

			Expect(resultConfigurations).To(HaveLen(1))
			uniqueID := testutils.UniqueIDFromObjectName(connectionPool, "public", "testconfiguration", backup.TYPE_TSCONFIGURATION)
			resultMetadata := resultMetadataMap[uniqueID]
			structmatcher.ExpectStructsToMatchExcluding(&configuration, &resultConfigurations[0], "Oid")
			structmatcher.ExpectStructsToMatch(&configurationMetadata, &resultMetadata)
		})
	})
})
