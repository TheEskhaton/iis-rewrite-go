package cmd

import (
	"encoding/xml"
	"os"

	"github.com/TheEskhaton/iis-toolkit/service"
	"github.com/spf13/cobra"
)

var (
	rewriteMapFile string
	writeFixToFile string
)

var checkCommand = &cobra.Command{
	Use:   "check",
	Short: "check a rewrite map for duplicates",
	Run: func(cmd *cobra.Command, args []string) {
		logger := service.NewLogger(false)
		logger.LogF("Checking rewrite map config %s for duplicates\n", rewriteMapFile)

		rewriteXml, err := service.NewRewriteMapRootXmlFromFile(rewriteMapFile, logger)
		if err != nil {
			logger.LogF("Error creating rewrite map from file : %v", err)
			return
		}
		outputXml := rewriteXml
		for i, rewriteMap := range rewriteXml.RewriteMap {
			rewriteKeys := make(map[string]string, 0)

			for _, mapping := range rewriteMap.Mappings {
				if _, ok := rewriteKeys[mapping.Key]; ok {
					logger.LogF("Duplicate key found: \"%v\" in map \"%v\"\n", mapping.Key, rewriteMap.Name)
					if writeFixToFile != "" {
						outputXml.RewriteMap[i].Mappings = removeMapping(rewriteXml.RewriteMap[i].Mappings, mapping.Key)
					}
				} else {
					rewriteKeys[mapping.Key] = mapping.Value
				}
			}
		}
		logger.LogLn("No other duplicates found")

		if writeFixToFile != "" {
			logger.LogLn("Fixing duplicates")
			output, err := xml.MarshalIndent(outputXml, "", "  ")
			if err != nil {
				logger.LogF("error marshalling line: %v\n", err)
			}
			err = os.Truncate(writeFixToFile, 0)
			if err != nil {
				logger.LogF("error truncating file: %v\n", err)
			}
			err = os.WriteFile(writeFixToFile, output, 0644)
			if err != nil {
				logger.LogF("error writing file: %v\n", err)
			}
		}
	},
}

func removeMapping(mappings []service.RewriteMapMappingXml, key string) []service.RewriteMapMappingXml {
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
	checkCommand.Flags().StringVarP(&writeFixToFile, "output", "o", "", "automatically fix duplicates and output to file")
}
