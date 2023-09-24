package main

import (
	"encoding/json"
	"facebook-business-sdk-codegen-relationships/database"
	"facebook-business-sdk-codegen-relationships/models"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
)

var (
	CodegenRepository  = "https://github.com/facebook/facebook-business-sdk-codegen"
	CodegenVersionFile = "/api_specs/specs/version.txt"
	CodegenLocalPath   = "sdk"
	ExcludedFiles      = []string{"version.txt", "enum_types.json"}
)

func PullCodegen() error {
	// verify if git is installed
	_, err := exec.LookPath("git")
	if err != nil {
		log.Fatal("git is not installed")
	}

	cmd := exec.Command("git", "clone", CodegenRepository, CodegenLocalPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func main() {
	fmt.Println("\n=============================================================")
	fmt.Println("         🚀 STARTING - CODEGEN RELATIONSHIPS IMPORTER			")
	fmt.Println("=============================================================\n")

	// Init Neo4j Driver
	database.InitNeo4j()

	// Reset All AdObjects & Fields
	err := database.ResetDatabase()
	if err != nil {
		log.Fatal(err)
	}

	// check if the sdk is already downloaded
	if _, err := os.Stat(CodegenLocalPath); os.IsNotExist(err) {
		errP := PullCodegen()
		if errP != nil {
			log.Fatal(errP)
		}
	}

	// check if the sdk is the sdk version in sdk/api_specs/specs/version.txt
	versionFile, errV := os.ReadFile(CodegenLocalPath + CodegenVersionFile)
	if errV != nil {
		log.Fatal(errV)
	}

	version := strings.Trim(string(versionFile), "\n")

	fmt.Println("Actual SDK version: " + version + "\n")

	// Read all json files in the api_specs/specs folder
	fileNames, err := os.ReadDir(CodegenLocalPath + "/api_specs/specs")
	if err != nil {
		log.Fatal(err)
	}

	parsed := make(map[string]models.AdObject)

	for _, name := range fileNames {
		// skip the excluded files
		if slices.Contains(ExcludedFiles, name.Name()) {
			continue
		}

		file, errJ := os.ReadFile(CodegenLocalPath + "/api_specs/specs/" + name.Name())
		if errJ != nil {
			log.Fatal(errJ)
		}

		var result models.AdObject
		err = json.Unmarshal(file, &result)

		parsed[strings.Replace(name.Name(), ".json", "", -1)] = result
	}

	// First Create all AdObjects with their fields
	for name, obj := range parsed {
		err := database.CreateAdObject(name, obj.Fields)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Then Create all Field Relationships
	for r, fields := range parsed {

		foundRelationship := false

		for _, field := range fields.Fields {

			toCheck := field.Type

			// example Type: AdAssetFeedSpecTitle or list<AdAssetFeedSpecTitle>
			re := regexp.MustCompile(`list<([^>]+)>`)
			match := re.FindAllStringSubmatch(toCheck, -1)

			if len(match) > 0 {
				toCheck = match[0][1]
			}

			if len(parsed[toCheck].Fields) > 0 {
				foundRelationship = true
				fmt.Println("🔥 Found relationship " + r + " ( " + field.Name + " ) <-> " + toCheck + "")

				err := database.CreateLinkBetweenFieldAndObject(toCheck)
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		if foundRelationship {
			fmt.Println("\n--------------------------------------\n")
		}
	}

	fmt.Println("\n✅ Importation Done")
}
