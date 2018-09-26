package backup

/*
 * This file contains structs and functions related to backing up metadata on the
 * master for objects relating to built-in text search that needs to be restored
 * before data is restored.
 *
 * Text search is not supported in GPDB 4.3, so none of these structs or functions
 * are used in a 4.3 backup.
 */

import (
	"sort"
	"strings"

	"github.com/greenplum-db/gpbackup/utils"
)

func PrintCreateTextSearchParserStatement(metadataFile *utils.FileWithByteCount, toc *utils.TOC, parser TextSearchParser, parserMetadata ObjectMetadata) {
	start := metadataFile.ByteCount
	parserFQN := utils.MakeFQN(parser.Schema, parser.Name)
	metadataFile.MustPrintf("\n\nCREATE TEXT SEARCH PARSER %s (", parserFQN)
	metadataFile.MustPrintf("\n\tSTART = %s,", parser.StartFunc)
	metadataFile.MustPrintf("\n\tGETTOKEN = %s,", parser.TokenFunc)
	metadataFile.MustPrintf("\n\tEND = %s,", parser.EndFunc)
	metadataFile.MustPrintf("\n\tLEXTYPES = %s", parser.LexTypesFunc)
	if parser.HeadlineFunc != "" {
		metadataFile.MustPrintf(",\n\tHEADLINE = %s", parser.HeadlineFunc)
	}
	metadataFile.MustPrintf("\n);")
	PrintObjectMetadata(metadataFile, parserMetadata, parserFQN, "TEXT SEARCH PARSER")
	toc.AddPredataEntry(parser.Schema, parser.Name, "TEXT SEARCH PARSER", "", start, metadataFile)
}

func PrintCreateTextSearchTemplateStatement(metadataFile *utils.FileWithByteCount, toc *utils.TOC, template TextSearchTemplate, templateMetadata ObjectMetadata) {
	start := metadataFile.ByteCount
	templateFQN := utils.MakeFQN(template.Schema, template.Name)
	metadataFile.MustPrintf("\n\nCREATE TEXT SEARCH TEMPLATE %s (", templateFQN)
	if template.InitFunc != "" {
		metadataFile.MustPrintf("\n\tINIT = %s,", template.InitFunc)
	}
	metadataFile.MustPrintf("\n\tLEXIZE = %s", template.LexizeFunc)
	metadataFile.MustPrintf("\n);")
	PrintObjectMetadata(metadataFile, templateMetadata, templateFQN, "TEXT SEARCH TEMPLATE")
	toc.AddPredataEntry(template.Schema, template.Name, "TEXT SEARCH TEMPLATE", "", start, metadataFile)
}

func PrintCreateTextSearchDictionaryStatement(metadataFile *utils.FileWithByteCount, toc *utils.TOC, dictionary TextSearchDictionary, dictionaryMetadata ObjectMetadata) {
	dictionaryFQN := utils.MakeFQN(dictionary.Schema, dictionary.Name)
	start := metadataFile.ByteCount
	metadataFile.MustPrintf("\n\nCREATE TEXT SEARCH DICTIONARY %s (", dictionaryFQN)
	metadataFile.MustPrintf("\n\tTEMPLATE = %s", dictionary.Template)
	if dictionary.InitOption != "" {
		metadataFile.MustPrintf(",\n\t%s", dictionary.InitOption)
	}
	metadataFile.MustPrintf("\n);")
	PrintObjectMetadata(metadataFile, dictionaryMetadata, dictionaryFQN, "TEXT SEARCH DICTIONARY")
	toc.AddPredataEntry(dictionary.Schema, dictionary.Name, "TEXT SEARCH DICTIONARY", "", start, metadataFile)
}

func PrintCreateTextSearchConfigurationStatement(metadataFile *utils.FileWithByteCount, toc *utils.TOC, configuration TextSearchConfiguration, configurationMetadata ObjectMetadata) {
	configurationFQN := utils.MakeFQN(configuration.Schema, configuration.Name)
	start := metadataFile.ByteCount
	metadataFile.MustPrintf("\n\nCREATE TEXT SEARCH CONFIGURATION %s (", configurationFQN)
	metadataFile.MustPrintf("\n\tPARSER = %s", configuration.Parser)
	metadataFile.MustPrintf("\n);")
	tokens := []string{}
	for token := range configuration.TokenToDicts {
		tokens = append(tokens, token)
	}
	sort.Strings(tokens)
	for _, token := range tokens {
		dicts := configuration.TokenToDicts[token]
		metadataFile.MustPrintf("\n\nALTER TEXT SEARCH CONFIGURATION %s", configurationFQN)
		metadataFile.MustPrintf("\n\tADD MAPPING FOR \"%s\" WITH %s;", token, strings.Join(dicts, ", "))
	}
	PrintObjectMetadata(metadataFile, configurationMetadata, configurationFQN, "TEXT SEARCH CONFIGURATION")
	toc.AddPredataEntry(configuration.Schema, configuration.Name, "TEXT SEARCH CONFIGURATION", "", start, metadataFile)
}
