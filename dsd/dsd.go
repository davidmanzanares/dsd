package main

import (
	"fmt"
	"log"
	"os"

	"github.com/davidmanzanares/dsd/dsdl"
	"github.com/spf13/cobra"
)

func main() {
	log.SetFlags(log.Lshortfile)
	conf, _ := dsdl.LoadConfig()

	cmdAdd := &cobra.Command{
		Use:   "add <target> <service> <pattern1> [patterns2]...",
		Short: "Add a new target to deploy",
		Long:  `Adds a new target to deploy, a target is composed by its name, its service URL and a list of glob patterns.`,
		Args:  cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			target := dsdl.Target{Name: args[0], Service: args[1], Patterns: args[2:]}
			err := dsdl.AddTarget(target)
			if err != nil {
				log.Println(err)
			}
			fmt.Printf("Target %s added\n", target)
		},
	}
	rootCmd := &cobra.Command{Use: "dsd <command>"}
	rootCmd.AddCommand(cmdAdd)

	cmdDeploy := &cobra.Command{
		Use:   "deploy <target>",
		Short: "Deploys to <target>",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			target, ok := conf.Targets[args[0]]
			if !ok {
				fmt.Printf("Target \"%s\" doesn't exist\n", args[0])
				os.Exit(1)
			}
			fmt.Println("Deploying to", target)
			v, err := dsdl.Deploy(*target)
			fmt.Println("Deployed ", v)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	rootCmd.AddCommand(cmdDeploy)

	cmdDownload := &cobra.Command{
		Use:   "download <service>",
		Short: "Downloads the current deployment on <service>",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := dsdl.Download(args[0])
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	rootCmd.AddCommand(cmdDownload)

	cmdWatch := &cobra.Command{
		Use:   "watch <service> [arg1] [arg2]...",
		Short: "Get <service> deployments, deploying the existing and new deployments",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dsdl.Watch(args[0], args[1:])
			select {}
		},
	}
	rootCmd.AddCommand(cmdWatch)

	rootCmd.Execute()
}
