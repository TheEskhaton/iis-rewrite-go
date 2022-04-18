package cmd

import (
	"encoding/xml"
	"os"

	"github.com/TheEskhaton/iis-rewrite-go/service"
	"github.com/spf13/cobra"
)

var rewriteMapFile string
var rewriteKeys = make(map[string]string, 0)

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

		for _, rewriteMap := range rewriteXml.RewriteMap {
			for _, mapping := range rewriteMap.Mappings {
				if _, ok := rewriteKeys[mapping.Key]; ok {
					logger.LogF("Duplicate key found: \"%v\" in map \"%v\"\n", mapping.Key, rewriteMap.Name)
				} else {
					rewriteKeys[mapping.Key] = mapping.Value
				}
			}
		}
		logger.LogLn("No other duplicates found")
	},
}

func init() {
	rootCmd.AddCommand(checkCommand)

	checkCommand.Flags().StringVarP(&rewriteMapFile, "file", "f", "", "rewrite map .config file")
}
