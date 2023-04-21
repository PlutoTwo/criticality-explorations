package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jupiterone/jupiterone-client-go/jupiterone"
	j1Domain "github.com/jupiterone/jupiterone-client-go/jupiterone/domain"
)

// MODIFY THIS
var assetNames = []string{"Device", "CodeRepo"}
var weights = &Weights{
	AgeInDays:    10,
	Dependencies: 10,
}

func areWeightsValid(weights *Weights) bool {
	if weights == nil {
		return false
	}

	if weights.AgeInDays < 0 || weights.AgeInDays > 10 {
		return false
	}
	if weights.Dependencies < 0 || weights.Dependencies > 10 {
		return false
	}

	return true
}

func getCriticalityForAsset(o GetCriticalityForAssetsOptions) {
	fmt.Printf("acquiring data for: %v\n", o.AssetName)
	criticalityPackages := []CriticalityPackage{}

	var query string = fmt.Sprintf(`
		FIND %s AS asset 
		THAT RELATES TO as relationships *
		RETURN 
			asset._id,
			asset.displayName,
			asset._class,
			asset._createdOn,
			count(relationships) as deps`,
		o.AssetName)

	qi := jupiterone.QueryInput{
		Query: query,
	}
	dependencyDataRaw, err := o.Config.JupiterOneClient.Query.Query(qi)
	if err != nil {
		fmt.Printf("error when retrieving results for %s: %v\n", o.AssetName, err.Error())
		o.ResultsChannel <- criticalityPackages
		return
	}

	resultsAsMap, ok := dependencyDataRaw.(map[string]interface{})
	if !ok {
		fmt.Printf("failed to convert results to map[string]interface{} for %s\n", o.AssetName)
		o.ResultsChannel <- criticalityPackages
		return
	}

	data, ok := resultsAsMap["data"].([]interface{})
	if !ok {
		fmt.Printf("failed to convert results to []interface{} for %s\n", o.AssetName)
		o.ResultsChannel <- criticalityPackages
		return
	}

	fmt.Printf("for %s, the number of data points found is: %v\n", o.AssetName, len(data))

	for _, depRaw := range data {
		jsonString, err := json.Marshal(depRaw)
		if err != nil {
			continue
		}
		dep := QueryDataReturnClause{}
		err = json.Unmarshal(jsonString, &dep)
		if err != nil {
			continue
		}

		// must have a _id value or else we can't update the entity
		if dep.ID == "" {
			continue
		}

		currentTime := time.Now()
		t, err := time.Parse(time.RFC3339, dep.CreatedOn)
		if err != nil {
			continue
		}
		daysSinceFloat := currentTime.Sub(t).Hours() / 24
		daysSince := int(daysSinceFloat)
		criticalityScore := ((dep.Dependencies * o.Weights.Dependencies) * (daysSince * o.Weights.AgeInDays)) / 10

		resultsFormat := CriticalityPackage{
			ID:               dep.ID,
			AssetName:        dep.DisplayName,
			Class:            dep.Class,
			Dependencies:     dep.Dependencies,
			AgeInDays:        daysSince,
			CriticalityScore: criticalityScore,
		}

		criticalityPackages = append(criticalityPackages, resultsFormat)
	}

	o.ResultsChannel <- criticalityPackages
	fmt.Printf("done acquiring data for: %v\n", o.AssetName)
}

func getJupiterOneClient(c *jupiterone.Config) *jupiterone.Client {
	jupiteroneClient, err := jupiterone.NewClient(c)
	if err != nil {
		fmt.Printf("unable to create a jupiterone client. error: %v", err.Error())
		os.Exit(1)
	}

	return jupiteroneClient
}

func getConfig() Config {
	configurationKeys := []string{"J1_ACCOUNT", "J1_API_KEY", "J1_REGION"}
	for _, key := range configurationKeys {
		if os.Getenv(key) == "" {
			fmt.Fprintf(os.Stderr, "missing configuration key: %v", key)
			os.Exit(1)
		}
	}

	jupiteroneConfig := jupiterone.Config{
		AccountID: os.Getenv("J1_ACCOUNT"),
		APIKey:    os.Getenv("J1_API_KEY"),
		Region:    os.Getenv("J1_REGION"),
	}

	return Config{
		JupiterOneConfig: jupiteroneConfig,
		JupiterOneClient: getJupiterOneClient(&jupiteroneConfig),
	}
}

func main() {
	fmt.Println("begin acquiring criticality data...")

	config := getConfig()
	weightsAreValid := areWeightsValid(weights)
	if !weightsAreValid {
		fmt.Printf("invalid weights provided. please provide weights between 0 and 10. provided weights: %v\n", weights)
		fmt.Print("exiting...")
		return
	}

	numberOfAssets := len(assetNames)
	jobs := make(chan string, numberOfAssets)
	resultsChannel := make(chan []CriticalityPackage, numberOfAssets)

	fmt.Printf("number of assets: %v\n", numberOfAssets)

	if numberOfAssets == 0 {
		fmt.Print("no assets specified.\nexiting...")
		return
	}
	if numberOfAssets > 5 {
		fmt.Printf("more than 5 assets specified (%v), please reduce the number of assets to acquire data for.\nexiting...", numberOfAssets)
		return
	}

	worker := func(id int, jobs <-chan string, resultsChannel chan []CriticalityPackage) {
		for assetName := range jobs {
			getCriticalityForAsset(GetCriticalityForAssetsOptions{
				Context:        context.Background(),
				Config:         config,
				AssetName:      assetName,
				Weights:        *weights,
				ResultsChannel: resultsChannel,
			})
		}

	}

	for w := 1; w <= numberOfAssets; w++ {
		go worker(w, jobs, resultsChannel)
	}

	for _, assetName := range assetNames {
		jobs <- assetName
	}

	close(jobs)

	var output []CriticalityPackage = []CriticalityPackage{}

	for a := 1; a <= numberOfAssets; a++ {
		outputfromAnAsset := <-resultsChannel
		output = append(output, outputfromAnAsset...)
	}

	if len(output) == 0 {
		fmt.Printf("no criticality data was found.\nexiting...")
		return
	}

	var uploadPayload j1Domain.SyncPayload = j1Domain.SyncPayload{}

	for _, item := range output {
		uploadPayloadItem := GetCriticalityEntityPayload{
			ID:               item.ID,
			CriticalityScore: item.CriticalityScore,
		}
		uploadPayload.Entities = append(uploadPayload.Entities, uploadPayloadItem)
	}

	fmt.Println("ingesting criticality data into JupiterOne...")
	startParams := j1Domain.StartParams{
		Source:   "api",
		SyncMode: "CREATE_OR_UPDATE",
	}
	config.JupiterOneClient.Synchronization.ProcessSyncJob(startParams, uploadPayload)

	fmt.Println("criticality data ingested into JupiterOne...")
}
