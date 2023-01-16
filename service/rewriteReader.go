package service

import (
	"encoding/xml"
	"os"
)

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

func NewRewriteMapRootXmlFromFile(rewriteMapFile string, logger Logger) (RewriteMapRootXml, error) {
	contents, err := os.ReadFile(rewriteMapFile)

	if err != nil {
		logger.LogF("Error reading file: %v\n", err)
	}

	rewriteXml := RewriteMapRootXml{}

	err = xml.Unmarshal([]byte(contents), &rewriteXml)
	if err != nil {
		logger.LogF("Error unmarshalling line: %v\n", err)
	}
	return rewriteXml, err
}
