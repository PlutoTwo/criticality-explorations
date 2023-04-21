package main

import (
	"context"

	"github.com/jupiterone/jupiterone-client-go/jupiterone"
)

type CriticalityPackage struct {
	ID               string      `json:"_id"`
	AssetName        string      `json:"assetName"`
	Class            interface{} `json:"assetClass"`
	Dependencies     int         `json:"dependencies"`
	AgeInDays        int         `json:"ageInDays"`
	CriticalityScore int         `json:"criticalityScore"`
}

type Config struct {
	JupiterOneConfig jupiterone.Config
	JupiterOneClient *jupiterone.Client
}

type QueryDataReturnClause struct {
	ID           string      `json:"asset._id"`
	DisplayName  string      `json:"asset.displayName"`
	Class        interface{} `json:"asset._class"`
	CreatedOn    string      `json:"asset._createdOn"`
	Dependencies int         `json:"deps"`
}

type GetCriticalityOutputResult struct {
	AssetName        string      `json:"assetName"`
	Class            interface{} `json:"assetClass"`
	Dependencies     int         `json:"dependencies"`
	AgeInDays        int         `json:"ageInDays"`
	CriticalityScore int         `json:"criticalityScore"`
}

type GetCriticalityEntityPayload struct {
	ID               string `json:"_id"`
	CriticalityScore int    `json:"criticalityScore"`
}

type GetCriticalityOutput struct {
	Results []GetCriticalityOutputResult `json:"results"`
}

type Weights struct {
	AgeInDays    int `json:"ageInDays"`
	Dependencies int `json:"dependencies"`
}

type GetCriticalityRequestBody struct {
	Assets  []string `json:"assets"`
	Weights *Weights `json:"weights,omitempty"`
}

type GetCriticalityForAssetsOptions struct {
	Context        context.Context
	Config         Config
	AssetName      string
	ResultsChannel chan []CriticalityPackage
	Weights        Weights
}
