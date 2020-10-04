package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/spf13/cobra"
)

func getAvailableProjects() []string {
	// Fetch available project names by looking at the place
	// where we have stored all the json files
	fmt.Println("Looking for availbale projects")
	home, _ := os.UserHomeDir()
	gcDataPath := filepath.Join(home, ".gc_data")

	projects := []string{}

	files, err := ioutil.ReadDir(gcDataPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		projects = append(projects, strings.TrimSuffix(f.Name(), filepath.Ext(f.Name())))
	}
	return projects
}

func checkIfFoundPreviously(previousMatch bool, str string, substr string) bool {
	// Returns true if at least once it matched
	return (previousMatch || strings.Contains(str, substr))
}

func readFromJSONAndDisplay(project string, f string, q string) {
	// if q is empty, display all. Otherwise
	// display only what is matching it

	ShowSSH, _ := rootCmd.Flags().GetBool("ssh")

	type AccessConfigs struct {
		Natip string `json:"natIP"`
	}
	type NetworkInterfaces struct {
		Network       string          `json:"network"`
		NetworkIP     string          `json:"networkIP"`
		Subnet        string          `json:"subnetwork"`
		AccessConfigs []AccessConfigs `json:"accessConfigs"`
	}
	type Tags struct {
		Items []string `json:"items"`
	}
	type VM struct {
		Name              string              `json:"name"`
		Status            string              `json:"status"`
		NetworkInterfaces []NetworkInterfaces `json:"networkInterfaces"`
		Tags              Tags                `json:"tags"`
		Zone              string              `json:"zone"`
	}

	jsonFile, err := os.Open(f)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	data, _ := ioutil.ReadAll(jsonFile)

	var result []VM

	json.Unmarshal([]byte(data), &result)

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 20, 8, 0, '\t', tabwriter.AlignRight)

	defer w.Flush()

	for _, i := range result {
		var externalAddrArr []string
		var internalAddrArr []string
		var networksArr []string

		zoneParts := strings.Split(i.Zone, "/")
		zone := zoneParts[len(zoneParts)-1]

		for _, n := range i.NetworkInterfaces {
			internalAddrArr = append(internalAddrArr, n.NetworkIP)
			parts := strings.Split(n.Network, "/")
			nw := parts[len(parts)-1]
			networksArr = append(networksArr, nw)
			for _, a := range n.AccessConfigs {
				externalAddrArr = append(externalAddrArr, a.Natip)
			}
		}

		tags := strings.Join(i.Tags.Items[:], ",")
		externalAddresses := strings.Join(externalAddrArr[:], ",")
		internalAddresses := strings.Join(internalAddrArr[:], ",")
		networks := strings.Join(networksArr[:], ",")

		matched := fuzzy.Find(q, []string{i.Name, i.Status, networks, internalAddresses, externalAddresses})
		if len(matched) > 0 {
			if ShowSSH {
				fmt.Printf("gcloud compute ssh %s@%s --project %s --zone %s\n", os.Getenv("USER"), i.Name, project, zone)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", i.Name, i.Status, networks, internalAddresses, externalAddresses, zone, tags)
			}
		}
	}
}

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:     "ls",
	Short:   "list VMs",
	Aliases: []string{"l"},
	Long: `List VMs in a project

Example 1:
==========
Command: gc ls infra
Explanation : This does a fuzzy search for project with "infra" and lists all the VMs
in it.

Example 2:
=========
Command: gc ls infra db1
Explanation: Fuzzy finds the project name and then "db1" in the name, internal/external IP address`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {

		var query string
		var matches []string

		projectString := args[0]

		availableProjects := getAvailableProjects()

		home, _ := os.UserHomeDir()

		matches = fuzzy.Find(projectString, availableProjects)

		if len(matches) == 0 {
			fmt.Println("Did not find any matching projects")
			fmt.Println("Maybe add it to the config and run an update?")
			os.Exit(0)
		}
		fmt.Printf("Checking in %d matching projects\n", len(matches))

		if len(args) == 2 {
			query = args[1]
		}

		for _, project := range matches {
			fmt.Printf("Project: %s, Query: %s\n", project, query)
			jsonPath := project + ".json"
			projInfoFile := filepath.Join(home, ".gc_data", jsonPath)
			fmt.Println(projInfoFile)
			readFromJSONAndDisplay(project, projInfoFile, query)
		}
	},
}

func init() {
	var SSH bool
	rootCmd.AddCommand(lsCmd)
	rootCmd.PersistentFlags().BoolVarP(&SSH, "ssh", "s", false, "Show ssh command")
}
