package cmd

import (
	"encoding/xml"
	"os"

	"github.com/TheEskhaton/iis-rewrite-go/service"
	"github.com/spf13/cobra"
)

var rewriteMapFile string
var fix bool

type RewriteMapMappingXml struct {
	XMLName xml.Name `xml:"add"`
	Key     string   `xml:"key,attr"`
	Value   string   `xml:"value,attr"`
}

type RewriteMapMapsXml struct {
	XMLName  xml.Name               `xml:"rewriteMap"`
	Name     string                 `xml:"name,attr"`
	Mappings []RewriteMapMappingXml `xml:"add"`
}

type RewriteMapRootXml struct {
	XMLName    xml.Name            `xml:"rewriteMaps"`
	RewriteMap []RewriteMapMapsXml `xml:"rewriteMap"`
}

var checkCommand = &cobra.Command{
	Use:   "check",
	Short: "check a rewrite map for duplicates",
	Run: func(cmd *cobra.Command, args []string) {
		logger := service.NewLogger(false)
		logger.LogF("Checking rewrite map config %s for duplicates\n", rewriteMapFile)

		contents, err := os.ReadFile(rewriteMapFile)

		if err != nil {
			logger.LogF("Error reading file: %v\n", err)
		}

		rewriteXml := RewriteMapRootXml{}

		err = xml.Unmarshal([]byte(contents), &rewriteXml)
		if err != nil {
			logger.LogF("Error unmarshalling line: %v\n", err)
		}
		outputXml := rewriteXml
		for i, rewriteMap := range rewriteXml.RewriteMap {
			var rewriteKeys = make(map[string]string, 0)
			var mappingsCopy = make([]RewriteMapMappingXml, len(rewriteMap.Mappings))
			copy(mappingsCopy, rewriteMap.Mappings)
			for _, mapping := range mappingsCopy {
				if _, ok := rewriteKeys[mapping.Key]; ok {
					logger.LogF("Duplicate key found: \"%v\" in map \"%v\"\n", mapping.Key, rewriteMap.Name)
					if fix {
						outputXml.RewriteMap[i].Mappings = removeMapping(rewriteXml.RewriteMap[i].Mappings, mapping.Key)
					}
				} else {
					rewriteKeys[mapping.Key] = mapping.Value
				}
			}
		}
		logger.LogLn("No other duplicates found")

		if fix {
			logger.LogLn("Fixing duplicates")
			output, err := xml.MarshalIndent(outputXml, "", "  ")
			if err != nil {
				logger.LogF("Error marshalling line: %v\n", err)
			}
			err = os.Truncate(rewriteMapFile, 0)
			if err != nil {
				logger.LogF("Error truncating file: %v\n", err)
			}
			err = os.WriteFile(rewriteMapFile, output, 0644)
			if err != nil {
				logger.LogF("Error writing file: %v\n", err)
			}
		}
	},
}

func removeMapping(mappings []RewriteMapMappingXml, key string) []RewriteMapMappingXml {
	for i, mapping := range mappings {
		if mapping.Key == key {
			return append(mappings[:i], mappings[i+1:]...)
		}
	}
	return mappings
}

func init() {
	rootCmd.AddCommand(checkCommand)

	checkCommand.Flags().StringVarP(&rewriteMapFile, "file", "f", "", "rewrite map .config file")
	checkCommand.Flags().BoolVarP(&fix, "fix", "x", false, "automatically fix duplicates in source file")
}
