package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

var rootCmd = &cobra.Command{
    Use:   "muxi",
    Short: "MUXI CLI - AI agent registry and deployment tool",
}

var loginCmd = &cobra.Command{
    Use:   "login",
    Short: "Log in to a registry",
    Run: func(cmd *cobra.Command, args []string) {
        registry := viper.GetString("registry")
        var token string

        fmt.Printf("Enter token for %s: ", registry)
        fmt.Scanln(&token)

        err := keyring.Set("muxi", registry, token)
        if err != nil {
            log.Fatalf("Failed to store token: %v", err)
        }

        fmt.Println("Token stored securely.")
    },
}

var whoamiCmd = &cobra.Command{
    Use:   "whoami",
    Short: "Show current registry token status",
    Run: func(cmd *cobra.Command, args []string) {
        registry := viper.GetString("registry")
        _, err := keyring.Get("muxi", registry)
        if err != nil {
            fmt.Printf("Not logged in to %s\n", registry)
        } else {
            fmt.Printf("Logged in to %s\n", registry)
        }
    },
}

var pushCmd = &cobra.Command{
    Use:   "push",
    Short: "Push a schema to the registry",
    Run: func(cmd *cobra.Command, args []string) {
        registry := viper.GetString("registry")
        token, err := keyring.Get("muxi", registry)
        if err != nil {
            log.Fatalf("You must login first using 'muxi login'\n")
        }
        fmt.Printf("Using token: %s\n", token) // in real CLI, do NOT print token
        fmt.Println("Pushing schema (placeholder)...")
    },
}

func init() {
    cobra.OnInitialize(initConfig)

    rootCmd.PersistentFlags().String("registry", "muxihub.com", "Muxi registry to use")
    viper.BindPFlag("registry", rootCmd.PersistentFlags().Lookup("registry"))

    rootCmd.AddCommand(loginCmd)
    rootCmd.AddCommand(whoamiCmd)
    rootCmd.AddCommand(pushCmd)
}

func initConfig() {
    viper.SetConfigName("config")
    viper.AddConfigPath("$HOME/.muxi")
    viper.AutomaticEnv()
    _ = viper.ReadInConfig()
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
    }
}
