# Criticality Explorations

Welcome to the Criticality Explorations project by PlutoTwo. By using this script, you can ingest a computed "criticality score" for the assets in your JupiterOne environment giving you an additional layer of insight and visibility into what exists in your digital inventory. The greatest benefit this gives your team is a starting point for knowing what to work on first. 

NOTE: To be clear, please note that this project is provided for your exploration only. It is provided AS-IS and PlutoTwo does not make any guarantees of any kind. Any issues discovered through the use of this project should be filed as GitHub Issues on the Repo. They will be reviewed as time permits, but there are no formal SLAs for support.

## Implications On Your JupiterOne Environment

This script *DOES NOT* create or delete entities and relationships in your JupiterOne environment. The only effect this will have on your JupiterOne environment is that a property titled `criticalityScore` will be added onto the assets that you have explicitly specified in this script. Removing that property is trivial and you are also free to ignore its existence entirely.

## Getting Started

1. [Install Go](https://go.dev/doc/install) on the machine that will be running the script

2. Modify lines 15 & 16 with your own values. 
    - For `assetNames`, valid values here are `_class` and `_type` values of JupiterOne entities. 
        - Example: `Device`, `CodeRepo`, `aws_account`
    - For `weights`, enter a value between `0` and `10` for each property. Altering these values will affect the computation of the criticality score.


3. Set environment variables. This script uses three environment variables:
```
$ export J1_ACCOUNT=
$ export J1_API_KEY=
$ export J1_REGION=us
```

4. Execute the script by running the following command: `go run main.go domain.go`

5. View the results in your JupiterOne environment. An example query that you can use as a starting point is: 

```
FIND * WITH criticalityScore != undefined
```
