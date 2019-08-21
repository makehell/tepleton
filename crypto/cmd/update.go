// Copyright © 2017 Ethan Frey
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Change the password for a private key",
	Long:  `Change the password for a private key.`,
	Run:   updatePassword,
}

func init() {
	RootCmd.AddCommand(updateCmd)
}

func updatePassword(cmd *cobra.Command, args []string) {
	if len(args) != 1 || len(args[0]) == 0 {
		fmt.Println("You must provide a name for the key")
		return
	}
	name := args[0]

	oldpass, err := getPassword("Enter the current passphrase:")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	newpass, err := getCheckPassword("Enter the new passphrase:", "Repeat the new passphrase:")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = Manager.Update(name, oldpass, newpass)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Password successfully updated!")
	}
}
