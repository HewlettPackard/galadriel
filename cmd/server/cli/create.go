package cli

import (
	"errors"
	"fmt"
	"github.com/HewlettPackard/galadriel/cmd/server/util"
	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var memberName string

var createCmd = &cobra.Command{
	Use:   "create <member | relationship>",
	Short: "Allows creation of members, memberships and relationships",
}

var createMemberCmd = &cobra.Command{
	Use:   "member <trust-domain> --name <name>",
	Args:  cobra.ExactArgs(1),
	Short: "Register a new member and return its associated access token.",

	RunE: func(cmd *cobra.Command, args []string) error {
		td := args[0]
		c := util.NewServerClient(defaultSocketPath)
		m, err := c.CreateMember(common.Member{TrustDomain: td})
		if err != nil {
			return err
		}

		fmt.Println("MemberID:", m.ID)
		fmt.Println("Token:", m.Tokens[0])
		return nil
	},
}

var createRelationshipCmd = &cobra.Command{
	Use:   "relationship <memA> <memB>",
	Short: "Register a new relationship.",
	Args:  cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("invalid operation. usage: create relationship <memA> <memB>")
		}

		idA, err := uuid.FromBytes([]byte(args[0]))
		idB, err := uuid.FromBytes([]byte(args[0]))
		if err != nil {
			return err
		}

		c := util.NewServerClient(defaultSocketPath)
		r, err := c.CreateRelationship(common.Relationship{
			MemberA: idA,
			MemberB: idB,
		})
		if err != nil {
			return err
		}

		fmt.Println("Relationship:", r)
		return nil
	},
}

func init() {
	createMemberCmd.Flags().StringVar(&memberName, "name", "", "member name")

	createCmd.AddCommand(createRelationshipCmd)
	createCmd.AddCommand(createMemberCmd)

	RootCmd.AddCommand(createCmd)
}
