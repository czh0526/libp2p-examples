package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "This is a sample application",
}

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Account related commands",
}

var friendCmd = &cobra.Command{
	Use:   "friend",
	Short: "Friend related commands",
}

func init() {
	// new account
	newAccountCmd.Flags().String("nick", "", "nick name")
	newAccountCmd.Flags().String("phone", "", "phone number")
	newAccountCmd.Flags().String("passphrase", "", "passphrase")

	// add friend
	addFriendCmd.Flags().String("id", "", "id of friend")
	addFriendCmd.Flags().String("nick", "", "nick name of friend")
	addFriendCmd.Flags().String("self-id", "", "the id of myself")

	// account commands
	accountCmd.AddCommand(newAccountCmd)

	// friend commands
	friendCmd.AddCommand(addFriendCmd)

	rootCmd.AddCommand(accountCmd)
	rootCmd.AddCommand(friendCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
