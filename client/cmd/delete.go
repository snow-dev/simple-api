/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	blogpb "github.com/snow-dev/simple-api/proto"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delte a blog post by its ID",
	Long:  `Delete a blog post by its mongoDB Unique identifier. If no blog post is found for de ID, it will return a 'Not found' error`,

	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := cmd.Flags().GetString("id")
		if err != nil {
			return err
		}

		req := &blogpb.DeleteBlogReq{
			Id: id,
		}
		// We only return true upon success for others cases an error is thrown
		// We can thus omit the response variable for now and just print something
		_, err = client.DeleteBlog(context.Background(), req)
		if err != nil {
			return err
		}

		fmt.Printf("Successfully deleted the blog with ID: %s\n", id)
		return nil
	},
}

func init() {
	deleteCmd.Flags().StringP("id", "i", "", "The id of the blog")
	deleteCmd.MarkFlagRequired("id")
	rootCmd.AddCommand(deleteCmd)
}
