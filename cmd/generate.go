package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/TheEskhaton/iis-toolkit/service"
	"github.com/spf13/cobra"
)

const rewriteRowTemplate string = "<add key=\"%v\" value=\"%v\" />\n"

var generatedMaps []rewriteMap

type rewriteMap struct {
	from string
	to   string
}

func (r rewriteMap) String() string {
	return fmt.Sprintf(rewriteRowTemplate, r.from, r.to)
}

type generateSettings struct {
	rewriteMapFile string
	rewireMapName  string
	separator      string
	silent         bool
	stripDomains   bool
	domainToRemove string
}

var generateConfig generateSettings

var generateCommand = &cobra.Command{
	Use:   "generate",
	Short: "generate a rewrite map",
	Run: func(cmd *cobra.Command, args []string) {
		logger := service.NewLogger(generateConfig.silent)

		logger.LogF("Generating rewrite map from %s named %v\n", generateConfig.rewriteMapFile, generateConfig.rewireMapName)
		if generateConfig.stripDomains {
			logger.LogF("Domain names will be removed during generation")
		}
		outputFile, err := os.Create("rewriteMap.config")
		if err != nil {
			logger.LogLn(fmt.Sprintf("Error opening file: %v", err))
		}

		defer outputFile.Close()

		var lines []string

		lines = append(lines, "<rewriteMaps>\n")
		lines = append(lines, fmt.Sprintf("\t<rewriteMap name=\"%v\">\n", generateConfig.rewireMapName))

		inputFile, err := os.Open(generateConfig.rewriteMapFile)
		if err != nil {
			logger.LogLn(fmt.Sprintf("Error opening file: %v", err))
		}
		defer inputFile.Close()

		csvReader := csv.NewReader(inputFile)
		for {
			// Read each record from csv
			csvLine, err := csvReader.Read()
			if err == io.EOF {
				break
			}

			if err != nil {
				logger.LogF("Error parsing line: %v\n", err)
			}
			if len(csvLine) != 2 {
				logger.LogF("Error parsing line: %s\n", csvLine)
				continue
			}
			fromUrl := strings.ReplaceAll(csvLine[0], "\"", "")
			toUrl := strings.ReplaceAll(csvLine[1], "\"", "")
			fromUrl = strings.ReplaceAll(fromUrl, "&", "&amp;")
			toUrl = strings.ReplaceAll(toUrl, "&", "&amp;")
			if generateConfig.stripDomains {
				parsedFromUrl, errFrom := url.Parse(fromUrl)
				parsedToUrl, errTo := url.Parse(toUrl)
				if errFrom != nil {
					logger.LogF("Error parsing URL: %s\n", fromUrl)
				}
				if errTo != nil {
					logger.LogF("Error parsing URL: %s\n", toUrl)
				}
				fromUrl = strings.ReplaceAll(fromUrl, parsedFromUrl.Host, "")
				toUrl = strings.ReplaceAll(toUrl, parsedToUrl.Host, "")
				fromUrl = strings.ReplaceAll(fromUrl, parsedFromUrl.Scheme, "")
				toUrl = strings.ReplaceAll(toUrl, parsedToUrl.Scheme, "")
				fromUrl = strings.ReplaceAll(fromUrl, "://", "")
				toUrl = strings.ReplaceAll(toUrl, "://", "")
				if generateConfig.domainToRemove != "" {
					fromUrl = strings.ReplaceAll(fromUrl, generateConfig.domainToRemove, "")
					toUrl = strings.ReplaceAll(toUrl, generateConfig.domainToRemove, "")
				}
			}
			rewriteMap := rewriteMap{from: fromUrl, to: toUrl}
			if rewriteMap.from == "" {
				logger.LogF("SKIP: Empty key: %v\n", rewriteMap)
				continue
			}
			for _, mapItem := range generatedMaps {
				if mapItem.from == rewriteMap.from {
					logger.LogF("SKIP: duplicate rewrite map item: %v\n", rewriteMap)
					continue
				}
			}

			generatedMaps = append(generatedMaps, rewriteMap)
			lines = append(lines, "\t\t"+rewriteMap.String())
		}
		lines = append(lines, "\t</rewriteMap>\n")
		lines = append(lines, "</rewriteMaps>\n")

		for _, line := range lines {
			writeLine(line, outputFile)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCommand)

	generateCommand.Flags().StringVarP(&generateConfig.rewriteMapFile, "file", "f", "", "rewrite map CSV file")
	generateCommand.Flags().StringVarP(&generateConfig.rewireMapName, "name", "n", "", "rewrite map name")
	generateCommand.Flags().StringVarP(&generateConfig.separator, "separator", "s", ",", "csv separator")
	generateCommand.Flags().BoolVarP(&generateConfig.silent, "silent", "q", false, "silent mode")
	generateCommand.Flags().BoolVarP(&generateConfig.stripDomains, "stripDomains", "d", false, "strip domain names")
	generateCommand.Flags().StringVarP(&generateConfig.domainToRemove, "domainToRemove", "r", "", "domain to remove")
}

func writeLine(line string, outputFile *os.File) {
	_, err := outputFile.WriteString(line)
	if err != nil {
		fmt.Println(fmt.Errorf("error writing to file: %v", err))
	}
}
