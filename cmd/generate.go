package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
}

var generateConfig generateSettings

var generateCommand = &cobra.Command{
	Use:   "generate",
	Short: "generate a rewrite map",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("Generating rewrite map from %s named %v\n", generateConfig.rewriteMapFile, generateConfig.rewireMapName)

		inputFile, err := os.Open(generateConfig.rewriteMapFile)
		defer inputFile.Close()
		if err != nil {
			fmt.Println(fmt.Errorf("Error opening file: %v", err))
		}

		outputFile, err := os.Create("rewriteMap.config")
		if err != nil {
			fmt.Println(fmt.Errorf("Error opening file: %v", err))
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

		scanner := bufio.NewScanner(inputFile)

		for scanner.Scan() {

			lineText := scanner.Text()

			splitLine := strings.Split(lineText, generateConfig.separator)
			if len(splitLine) != 2 {
				fmt.Printf("Error parsing line: %s\n", lineText)
				continue
			}

			rewriteMap := rewriteMap{from: splitLine[0], to: splitLine[1]}

			for _, mapItem := range generatedMaps {
				if mapItem.from == rewriteMap.from {
					fmt.Printf("SKIP: duplicate rewrite map item: %v\n", rewriteMap)
					continue
				}
			}

			generatedMaps = append(generatedMaps, rewriteMap)

			fmt.Printf(rewriteMap.String())
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
}

func writeLine(line string, outputFile *os.File) {
	_, err := outputFile.WriteString(line)
	if err != nil {
		fmt.Println(fmt.Errorf("Error writing to file: %v", err))
	}
}
