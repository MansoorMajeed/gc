package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func fetchGcloudAndWrite(project string) {
	fmt.Println("Fetching : ", project)
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Failed to fetch home directory")
	}
	gcStorage := filepath.Join(home, ".gc_data")
	os.MkdirAll(gcStorage, os.ModePerm)

	gcloudPath, err := exec.LookPath("gcloud")
	if err != nil {
		fmt.Println("Gcloud not installed")
	}

	comm := exec.Command(gcloudPath, "compute", "instances", "list", "--project", project, "--format=json")
	out, err := comm.Output()

	if err != nil {
		fmt.Println("Error fetching from Project: ", project)
		fmt.Println("Error: ", err)
	}

	projectDataFile := project + ".json"
	projectDataFileFull := filepath.Join(gcStorage, projectDataFile)
	f, err := os.Create(projectDataFileFull)
	check(err)

	defer f.Close()
	f.WriteString(string(out[:]))

	fmt.Println("[Done] : ", project)
}

func getAllProjects() []string {
	comm := "gcloud projects list --format=json"
	fmt.Println("Executing: ", comm)
	return []string{"foo1", "foo2"}
}

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update the offline storage",
	Aliases: []string{"u"},
	Long: `Talk to Google cloud and fetch the latst information
about the VMs. By default it updates all projects.
Update only a project by using --project`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Updating project information")

		// Read the projects from the config file
		proj, _ := rootCmd.Flags().GetString("project")
		projects := []string{}
		if proj == "all" {
			// Fetch all projects
			fmt.Println("Getting ALL projects. This is gonna be mad")
			projects = getAllProjects()
		} else if proj == "" {
			// Get the projects from the config
			fmt.Println("Project not passed. Fetching from config")
			projects = viper.GetStringSlice("projects")
		} else {
			// Update only the project passed with the arg
			projects = append(projects, proj)
		}

		fmt.Printf("There are %d projects. Gonna spawn that many goroutines..\n", len(projects))

		var wg sync.WaitGroup

		// We probably should not be doing this if we have a lots of projects, but YOLO!
		for _, p := range projects {
			wg.Add(1)

			go func(project string) {
				fetchGcloudAndWrite(project)
				wg.Done()
			}(p) // pass the "p" to the anonymous function
		}

		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
