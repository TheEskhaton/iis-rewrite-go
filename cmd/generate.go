package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/TheEskhaton/iis-toolkit/service"
	"github.com/spf13/cobra"
)

var rewriteRowTemplate string = "<add key=\"%v\" value=\"%v\" />\n"
var generatedMaps []rewriteMap = make([]rewriteMap, 0)

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
}

var generateConfig generateSettings

var generateCommand = &cobra.Command{
	Use:   "generate",
	Short: "generate a rewrite map",
	Run: func(cmd *cobra.Command, args []string) {
		logger := service.NewLogger(generateConfig.silent)

		logger.LogF("Generating rewrite map from %s named %v\n", generateConfig.rewriteMapFile, generateConfig.rewireMapName)

		outputFile, err := os.Create("rewriteMap.config")
		if err != nil {
			logger.LogLn(fmt.Sprintf("Error opening file: %v", err))
		}

		lines := make(chan string)

		go func() {
			for line := range lines {
				writeLine(line, outputFile)
			}
			outputFile.Close()
		}()

		lines <- "<rewriteMaps>\n"
		lines <- fmt.Sprintf("\t<rewriteMap name=\"%v\">\n", generateConfig.rewireMapName)

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

			rewriteMap := rewriteMap{from: strings.ReplaceAll(csvLine[0], "\"", ""), to: strings.ReplaceAll(csvLine[1], "\"", "")}
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
			lines <- "\t\t" + rewriteMap.String()
		}
		lines <- "\t</rewriteMap>\n"
		lines <- "</rewriteMaps>\n"
		close(lines)
	},
}

func init() {
	rootCmd.AddCommand(generateCommand)

	generateCommand.Flags().StringVarP(&generateConfig.rewriteMapFile, "file", "f", "", "rewrite map CSV file")
	generateCommand.Flags().StringVarP(&generateConfig.rewireMapName, "name", "n", "", "rewrite map name")
	generateCommand.Flags().StringVarP(&generateConfig.separator, "separator", "s", ",", "csv separator")
	generateCommand.Flags().BoolVarP(&generateConfig.silent, "silent", "q", false, "silent mode")
}

func writeLine(line string, outputFile *os.File) {
	_, err := outputFile.WriteString(line)
	if err != nil {
		fmt.Println(fmt.Errorf("Error writing to file: %v", err))
	}
}
