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
	rootCmd := &cobra.Command{Use: "dsd <command>"}

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

	cmdRun := &cobra.Command{
		Use:   "run [--hotreload] [--on-success <reaction>] [--on-failure <reaction>] <service>",
		Short: "Run the deployed application on the target service",
		Long: `Run the deployed application on <service>, with the provided arguments.` + "\n" +
			`Where <reaction> is one of "exit", "wait" or "restart".` + "\n\t" +
			`"exit" will stop dsd's execution` + "\n\t" + `"wait" will wait for future updates (which will trigger an application start)` + "\n\t" +
			`"restart" will restart the application immediately.`,

		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hotreload, _ := cmd.Flags().GetBool("hotreload")
			onSuccess, _ := cmd.Flags().GetString("on-success")
			onFailure, _ := cmd.Flags().GetString("on-failure")

			successReaction, err := getReaction(onSuccess)
			if err != nil {
				fmt.Println(err)
				return
			}
			failureReaction, err := getReaction(onFailure)
			if err != nil {
				fmt.Println(err)
				return
			}

			r, err := dsdl.Run(args[0], dsdl.RunConf{
				HotReload: hotreload,
				OnSuccess: successReaction,
				OnFailure: failureReaction,
				Args:      args[1:]})
			if err != nil {
				fmt.Println(err)
				return
			}

			for {
				ev := r.WaitForEvent()
				fmt.Println(ev)
				if ev.Type == dsdl.Stopped {
					return
				}
			}
		},
	}
	cmdRun.Flags().Bool("hotreload", false, "If set, the application will be stopped and restarted with future updates.")
	cmdRun.Flags().String("on-success", "exit", `Reaction to application exits with a zero code.`)
	cmdRun.Flags().String("on-failure", "exit", `Reaction to application exits with a non-zero code.`)
	rootCmd.AddCommand(cmdRun)

	rootCmd.Execute()
}

func getReaction(s string) (dsdl.RunReaction, error) {
	if s == "restart" {
		return dsdl.Restart, nil
	}
	if s == "wait" {
		return dsdl.Wait, nil
	}
	if s == "exit" {
		return dsdl.Exit, nil
	}
	return dsdl.Exit, fmt.Errorf(`Invalid reaction (%s). Valid values are: "restart", "wait", "exit".`, s)
}
