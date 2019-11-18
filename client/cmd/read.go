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

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Find a Blog post by its ID",
	Long:  `Find a blog post by it's mongoDB Unique identifier'`,

	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := cmd.Flags().GetString("id")
		if err != nil {
			return err
		}

		req := &blogpb.ReadBlogReq{
			Id: id,
		}
		res, err := client.ReadBlog(context.Background(), req)
		if err != nil {
			return err
		}
		fmt.Println(res)
		return nil
	},
}

func init() {
	readCmd.Flags().StringP("id", "i", "", "The id of the blog")
	readCmd.MarkFlagRequired("id")
	rootCmd.AddCommand(readCmd)
}
