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
	"io"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all blog posts",
	Long:  `List all blog posts on streaming.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		// Create the request
		req := &blogpb.ListBlogsReq{}
		// Call ListBlogs that returns a stream
		stream, err := client.ListBlogs(context.Background(), req)
		// Check for errors.
		if err != nil {
			return err
		}
		// Start iterating
		for {
			// stream.Recv returns a pointer to a ListBlogReq at the current iteration
			res, err := stream.Recv()
			// If end of stream, break the loop
			if err == io.EOF {
				break
			}
			// if err, return an error
			if err != nil {
				return err
			}
			// return empty if there are no blogs
			// If everything went well use the generated getter to print the blog message
			fmt.Printf("Response %v", res)
			fmt.Println(res.GetBlog())
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
