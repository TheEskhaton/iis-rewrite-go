package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/TheEskhaton/iis-rewrite-go/service"
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
		defer inputFile.Close()
		if err != nil {
			logger.LogLn(fmt.Sprintf("Error opening file: %v", err))
		}

		scanner := bufio.NewScanner(inputFile)

		for scanner.Scan() {

			lineText := scanner.Text()

			splitLine := strings.Split(lineText, generateConfig.separator)
			if len(splitLine) != 2 {
				logger.LogF("Error parsing line: %s\n", lineText)
				continue
			}

			rewriteMap := rewriteMap{from: splitLine[0], to: splitLine[1]}
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

			logger.LogLn(rewriteMap.String())
			lines <- "\t\t" + rewriteMap.String()
		}
		lines <- "\t</rewriteMap>\n"
		lines <- "</rewriteMaps>\n"
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
