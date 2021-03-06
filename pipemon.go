package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	_ "github.com/lib/pq"
  "github.com/spf13/viper"

  "github.com/guilhermeslk/pipemon/models"
)


const (
	jisPipelineName = "JIS"
	jpsPipelineName = "JPS"
	reloadInterval = 5
)

var (
	dbJerico *sql.DB
	dbJIS    *sql.DB
	dbJPS    *sql.DB
)

func main() {
  viper.SetConfigType("yaml")
  viper.SetConfigName("pipemon_database")
  viper.AddConfigPath("$GOPATH/bin")
  err := viper.ReadInConfig() // Find and read the config file
  if err != nil { // Handle errors reading the config file
      panic(fmt.Errorf("Fatal error config file: %s \n", err))
  }

  dbHost       := viper.GetString("host")
  dbUser       := viper.GetString("user")
  dbPassword   := viper.GetString("password")
  jericoDbName := viper.GetString("jerico")
  jisDbName    := viper.GetString("jis")
  jpsDbName    := viper.GetString("jps")

	dbJerico = models.InitDB(fmt.Sprintf("host=%s user=%s password=%s dbname=%s",dbHost, dbUser, dbPassword, jericoDbName))
	dbJIS = models.InitDB(fmt.Sprintf("host=%s user=%s password=%s dbname=%s",dbHost, dbUser, dbPassword, jisDbName))
	dbJPS = models.InitDB(fmt.Sprintf("host=%s user=%s password=%s dbname=%s",dbHost, dbUser, dbPassword, jpsDbName))

	listPipelines(dbJerico)
}

func listPipelines(db *sql.DB) {
	pipelines, err := models.QueryPipelines(db)
	checkErr(err)

	clearScr()
	fmt.Println(time.Now().Local().String())

	color.Yellow("### PIPEMON ###")

	for _, pipeline := range pipelines {
		printID(pipeline.Id)
		printSeparator()
		fmt.Printf(pipeline.Type)
		printSeparator()
		printState(pipeline.State)
		fmt.Printf("\n")
	}

	fmt.Print("Pipeline to query: ")

	var input int
	_, err = fmt.Scanf("%v\n", &input)
	checkErr(err)

	ticker := time.NewTicker(time.Second * reloadInterval)

	for range ticker.C {
		go showPipelineDetails(input, dbJerico)
	}
}

func showPipelineDetails(pipelineID int, db *sql.DB) {
	clearScr()

	fmt.Println(time.Now().Local().String())
	fmt.Printf("PIPELINE: %v", pipelineID)
	fmt.Printf("\n")

	listPipelineSteps(pipelineID, db, 0)
}

func listPipelineSteps(pipelineID int, db *sql.DB, paddingLength int) {
	steps, err := models.QueryPipelineSteps(pipelineID, db)
	checkErr(err)

	for _, step := range steps {
		printPipelineStep(step, paddingLength)

		json.Unmarshal([]byte(step.AsyncResult), &step.AsyncResultData)
		checkErr(err)

		if step.AsyncResultData["external_pipeline_id"] != nil &&
			step.AsyncResultData["external_pipeline_name"] != nil {
			externalPipelineID := step.AsyncResultData["external_pipeline_id"].(float64)
			externalPipelineName := step.AsyncResultData["external_pipeline_name"].(string)

			if externalPipelineName == jisPipelineName {
				listPipelineSteps(int(externalPipelineID), dbJIS, 5)
			} else if externalPipelineName == jpsPipelineName {
				listPipelineSteps(int(externalPipelineID), dbJPS, 5)
			}
		}
	}
}

func printPipelineStep(step *models.PipelineStep, paddingLength int) {
	printPadding(paddingLength)
	printID(step.Id)
	printSeparator()
	printStep(step.StepClass)
	printSeparator()
	printState(step.State)
	fmt.Printf("\n")
}

func printPadding(length int) {
	if length <= 0 {
		return
	}

	for i := 0; i < length; i++ {
		fmt.Printf(" ")
	}

	fmt.Printf("|")
}

func printID(id int) {
	white := color.New(color.FgWhite).PrintfFunc()
	white("%4v", id)
}

func printStep(stepClass string) {
	blue := color.New(color.FgBlue).PrintfFunc()
	blue("%-65s", stepClass)
}

func printSeparator() {
	fmt.Printf(" | ")
}

func printState(state string) {
	red := color.New(color.FgRed).PrintfFunc()
	green := color.New(color.FgGreen).PrintfFunc()
	cyan := color.New(color.FgCyan).PrintfFunc()

	if state == "running" {
		cyan("%v", strings.ToUpper(state))
	} else if state == "failed" {
		red("%v", strings.ToUpper(state))
	} else if state == "done" {
		green("%v", strings.ToUpper(state))
	} else {
		fmt.Printf("PENDING")
	}
}

func clearScr() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
