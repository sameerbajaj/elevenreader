package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sameerbajaj/elevenreader/internal/api"
	"github.com/sameerbajaj/elevenreader/internal/auth"
	"github.com/sameerbajaj/elevenreader/internal/config"
	"github.com/sameerbajaj/elevenreader/internal/ui"
)

var (
	version = "1.0.0"
	commit  = "none"
	date    = "unknown"

	flagToken   string
	flagAPIKey  string
	flagBaseURL string
	flagJSON    bool
)

func getClient() (*api.Client, error) {
	creds, err := auth.ResolveCredentials(flagToken, flagAPIKey, flagBaseURL)
	if err != nil {
		return nil, err
	}
	return api.NewClient(creds.BaseURL, creds.Token, creds.APIKey), nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "elevenreader",
		Short: "A fast, clean CLI for ElevenReader (https://elevenreader.io/reader/library)",
		Long: `elevenreader is a standalone command-line client for ElevenReader library CRUD operations.
Easily list, view, add, update, archive, and delete articles, books, documents, collections, and bookmarks.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "ElevenReader session token or Bearer JWT")
	rootCmd.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "ElevenLabs API key (xi-api-key)")
	rootCmd.PersistentFlags().StringVar(&flagBaseURL, "base-url", "", "Custom API base URL (default: https://api.elevenlabs.io/v1/reader)")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output results as JSON")

	// Library commands
	libraryCmd := newLibraryCmd()
	rootCmd.AddCommand(libraryCmd)

	// Direct top-level shortcuts for library CRUD
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newGetCmd())
	rootCmd.AddCommand(newReadCmd())
	rootCmd.AddCommand(newAddCmd())
	rootCmd.AddCommand(newUpdateCmd())
	rootCmd.AddCommand(newArchiveCmd())
	rootCmd.AddCommand(newUnarchiveCmd())
	rootCmd.AddCommand(newDeleteCmd())

	// Collections
	rootCmd.AddCommand(newCollectionCmd())

	// Bookmarks
	rootCmd.AddCommand(newBookmarkCmd())

	// Auth
	rootCmd.AddCommand(newAuthCmd())

	// Version
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the elevenreader CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			if flagJSON {
				_ = ui.PrintJSON(map[string]string{
					"version": version,
					"commit":  commit,
					"date":    date,
				})
				return
			}
			fmt.Printf("elevenreader version %s (commit: %s, built: %s)\n", version, commit, date)
		},
	})

	if err := rootCmd.Execute(); err != nil {
		ui.ErrorPrint("%s", err)
		os.Exit(1)
	}
}

func newLibraryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "library",
		Aliases: []string{"lib", "reads"},
		Short:   "Manage your ElevenReader library items",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, args)
		},
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newReadCmd())
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newArchiveCmd())
	cmd.AddCommand(newUnarchiveCmd())
	cmd.AddCommand(newDeleteCmd())

	return cmd
}

var (
	flagPageSize   int
	flagSortBy     string
	flagLastSortID string
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List reads in your ElevenReader library",
		RunE:    runList,
	}

	cmd.Flags().IntVar(&flagPageSize, "page-size", 30, "Maximum number of items to return")
	cmd.Flags().StringVar(&flagSortBy, "sort-by", "recently_added_desc", "Field to sort by (recently_added_desc, recently_added_asc, recently_listened_desc, recently_listened_asc)")
	cmd.Flags().StringVar(&flagLastSortID, "last-sort-id", "", "Cursor ID for pagination")
	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	res, err := client.ListReads(api.ListReadsOptions{
		PageSize:   flagPageSize,
		SortBy:     flagSortBy,
		LastSortID: flagLastSortID,
	})
	if err != nil {
		return err
	}

	if flagJSON {
		return ui.PrintJSON(res)
	}

	ui.PrintReadsTable(res.Reads)
	if res.HasMore && res.LastSortID != nil {
		fmt.Printf("\n(More items available. Pass --last-sort-id %s to view next page)\n", *res.LastSortID)
	}
	return nil
}

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "get <read-id>",
		Aliases: []string{"info", "show"},
		Short:   "Get details of a specific read item",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			read, err := client.GetRead(args[0])
			if err != nil {
				return err
			}

			if flagJSON {
				return ui.PrintJSON(read)
			}

			ui.PrintReadDetails(read)
			return nil
		},
	}
}

var (
	flagSource bool
	flagRaw    bool
	flagOutput string
)

func newReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "read <read-id>",
		Aliases: []string{"cat", "text", "markdown", "view"},
		Short:   "View or export the text/markdown content of a read",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			var content string
			if flagSource {
				content, err = client.GetReadSource(args[0])
			} else {
				content, err = client.GetReadMarkdown(args[0])
				if !flagRaw {
					content = ui.CleanMarkdown(content)
				}
			}
			if err != nil {
				return err
			}

			if flagOutput != "" {
				if err := os.WriteFile(flagOutput, []byte(content), 0644); err != nil {
					return fmt.Errorf("writing output file: %w", err)
				}
				ui.SuccessPrint("Saved content to %s (%d bytes)", flagOutput, len(content))
				return nil
			}

			fmt.Print(content)
			return nil
		},
	}

	cmd.Flags().BoolVar(&flagSource, "source", false, "Fetch raw source text instead of Markdown")
	cmd.Flags().BoolVar(&flagRaw, "raw", false, "Keep raw span/timing tags in Markdown")
	cmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Save content to a local file")
	return cmd
}

var (
	flagURL      string
	flagFile     string
	flagText     string
	flagTitle    string
	flagAuthor   string
	flagDesc     string
	flagUnread   bool
	flagYes      bool
)

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add",
		Aliases: []string{"import", "upload", "create"},
		Short:   "Add a website article, document, or text to your library",
		Example: `  elevenreader add --url https://en.wikipedia.org/wiki/Go_(programming_language)
  elevenreader add --file document.pdf --title "My Document"
  elevenreader add --text "Quick note content" --title "My Note"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagURL == "" && flagFile == "" && flagText == "" {
				return fmt.Errorf("must provide one of --url, --file, or --text")
			}

			client, err := getClient()
			if err != nil {
				return err
			}

			created, err := client.AddRead(api.AddReadOptions{
				SourceURL:   flagURL,
				FilePath:    flagFile,
				Text:        flagText,
				Title:       flagTitle,
				Author:      flagAuthor,
				Description: flagDesc,
			})
			if err != nil {
				return err
			}

			if flagJSON {
				return ui.PrintJSON(created)
			}

			title := created.Title
			if title == "" {
				title = "(untitled)"
			}
			ui.SuccessPrint("Successfully added '%s' to ElevenReader library!", title)
			fmt.Printf("Read ID: %s\n", created.ReadID)
			return nil
		},
	}

	cmd.Flags().StringVar(&flagURL, "url", "", "URL of article or webpage to import")
	cmd.Flags().StringVarP(&flagFile, "file", "f", "", "Path to local document (PDF, EPUB, TXT, etc.)")
	cmd.Flags().StringVar(&flagText, "text", "", "Raw text content to import")
	cmd.Flags().StringVar(&flagTitle, "title", "", "Custom title for the read")
	cmd.Flags().StringVar(&flagAuthor, "author", "", "Author name")
	cmd.Flags().StringVar(&flagDesc, "description", "", "Description")
	return cmd
}

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <read-id>",
		Aliases: []string{"edit", "set"},
		Short:   "Update metadata for an existing read item",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			payload := api.ReadUpdatePayload{}
			hasUpdate := false

			if cmd.Flags().Changed("title") {
				payload.Title = &flagTitle
				hasUpdate = true
			}
			if cmd.Flags().Changed("author") {
				payload.Author = &flagAuthor
				hasUpdate = true
			}
			if cmd.Flags().Changed("description") {
				payload.Description = &flagDesc
				hasUpdate = true
			}
			if cmd.Flags().Changed("unread") {
				payload.MarkedUnread = &flagUnread
				hasUpdate = true
			}

			if !hasUpdate {
				return fmt.Errorf("no fields specified to update (pass --title, --author, --description, or --unread)")
			}

			if err := client.UpdateRead(args[0], payload); err != nil {
				return err
			}

			ui.SuccessPrint("Updated read %s", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&flagTitle, "title", "", "Updated title")
	cmd.Flags().StringVar(&flagAuthor, "author", "", "Updated author")
	cmd.Flags().StringVar(&flagDesc, "description", "", "Updated description")
	cmd.Flags().BoolVar(&flagUnread, "unread", false, "Mark as unread")
	return cmd
}

func newArchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "archive <read-id>",
		Short: "Archive a read item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			if err := client.ArchiveRead(args[0]); err != nil {
				return err
			}

			ui.SuccessPrint("Archived read %s", args[0])
			return nil
		},
	}
}

func newUnarchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unarchive <read-id>",
		Short: "Unarchive a read item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			if err := client.UnarchiveRead(args[0]); err != nil {
				return err
			}

			ui.SuccessPrint("Unarchived read %s", args[0])
			return nil
		},
	}
}

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <read-id>",
		Aliases: []string{"rm"},
		Short:   "Permanently delete a read item from your library",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			readID := args[0]
			if !flagYes {
				fmt.Printf("Are you sure you want to permanently delete read %s? [y/N]: ", readID)
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(strings.ToLower(input))
				if input != "y" && input != "yes" {
					fmt.Println("Aborted.")
					return nil
				}
			}

			if err := client.DeleteRead(readID); err != nil {
				return err
			}

			ui.SuccessPrint("Deleted read %s", readID)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func newCollectionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "collection",
		Aliases: []string{"collections", "col"},
		Short:   "Manage custom collections/folders in ElevenReader",
	}

	// List
	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all user collections",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			cols, err := client.ListCollections()
			if err != nil {
				return err
			}

			if flagJSON {
				return ui.PrintJSON(cols)
			}

			ui.PrintCollectionsTable(cols)
			return nil
		},
	})

	// Create
	var colIcon string
	createCmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new collection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			col, err := client.CreateCollection(args[0], colIcon)
			if err != nil {
				return err
			}

			if flagJSON {
				return ui.PrintJSON(col)
			}

			ui.SuccessPrint("Created collection '%s' (ID: %s)", col.Title, col.ID)
			return nil
		},
	}
	createCmd.Flags().StringVar(&colIcon, "icon", "📚", "Emoji icon for the collection")
	cmd.AddCommand(createCmd)

	// Delete
	cmd.AddCommand(&cobra.Command{
		Use:     "delete <collection-id>",
		Aliases: []string{"rm"},
		Short:   "Delete a collection",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			if err := client.DeleteCollection(args[0]); err != nil {
				return err
			}

			ui.SuccessPrint("Deleted collection %s", args[0])
			return nil
		},
	})

	// Add read to collection
	cmd.AddCommand(&cobra.Command{
		Use:   "add <collection-id> <read-id>",
		Short: "Add a read item to a collection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			if err := client.AddReadToCollection(args[0], args[1]); err != nil {
				return err
			}

			ui.SuccessPrint("Added read %s to collection %s", args[1], args[0])
			return nil
		},
	})

	// Remove read from collection
	cmd.AddCommand(&cobra.Command{
		Use:   "remove <collection-id> <read-id>",
		Short: "Remove a read item from a collection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			if err := client.RemoveReadFromCollection(args[0], args[1]); err != nil {
				return err
			}

			ui.SuccessPrint("Removed read %s from collection %s", args[1], args[0])
			return nil
		},
	})

	return cmd
}

func newBookmarkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bookmark",
		Aliases: []string{"bookmarks", "bm"},
		Short:   "View and export bookmarks/highlights",
	}

	// List
	cmd.AddCommand(&cobra.Command{
		Use:     "list <read-id>",
		Aliases: []string{"ls"},
		Short:   "List bookmarks for a specific read",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			bms, err := client.ListBookmarks(args[0])
			if err != nil {
				return err
			}

			if flagJSON {
				return ui.PrintJSON(bms)
			}

			ui.PrintBookmarksTable(bms)
			return nil
		},
	})

	// Export
	var bmOutput string
	exportCmd := &cobra.Command{
		Use:   "export <read-id>",
		Short: "Export bookmarks as Markdown",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			md, err := client.ExportBookmarksMarkdown(args[0])
			if err != nil {
				return err
			}

			if bmOutput != "" {
				if err := os.WriteFile(bmOutput, []byte(md), 0644); err != nil {
					return err
				}
				ui.SuccessPrint("Exported bookmarks to %s", bmOutput)
				return nil
			}

			fmt.Print(md)
			return nil
		},
	}
	exportCmd.Flags().StringVarP(&bmOutput, "output", "o", "", "File to save exported bookmarks to")
	cmd.AddCommand(exportCmd)

	// Delete
	cmd.AddCommand(&cobra.Command{
		Use:     "delete <bookmark-id>",
		Aliases: []string{"rm"},
		Short:   "Delete a bookmark",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			if err := client.DeleteBookmark(args[0]); err != nil {
				return err
			}

			ui.SuccessPrint("Deleted bookmark %s", args[0])
			return nil
		},
	})

	return cmd
}

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage ElevenReader authentication and credentials",
	}

	// Status
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := auth.ResolveCredentials(flagToken, flagAPIKey, flagBaseURL)
			if err != nil {
				fmt.Println("Status: Unauthenticated")
				fmt.Printf("Reason: %s\n", err.Error())
				return nil
			}

			tokenVal := creds.Token
			if tokenVal == "" {
				tokenVal = creds.APIKey
			}

			if flagJSON {
				return ui.PrintJSON(map[string]string{
					"status":   "authenticated",
					"source":   creds.Source,
					"base_url": creds.BaseURL,
					"token":    auth.MaskToken(tokenVal),
				})
			}

			fmt.Println("Status:   Authenticated")
			fmt.Printf("Source:   %s\n", creds.Source)
			fmt.Printf("Base URL: %s\n", creds.BaseURL)
			fmt.Printf("Token:    %s\n", auth.MaskToken(tokenVal))
			return nil
		},
	})

	// Login
	var loginToken, loginKey string
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Store credentials in config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if loginToken == "" && loginKey == "" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter your ElevenReader session token or ElevenLabs API key: ")
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)
				if input == "" {
					return fmt.Errorf("no input provided")
				}
				if strings.HasPrefix(input, "sk_") {
					loginKey = input
				} else {
					loginToken = input
				}
			}

			cfg, _ := config.Load()
			if cfg == nil {
				cfg = &config.Config{}
			}

			if loginToken != "" {
				cfg.Token = loginToken
			}
			if loginKey != "" {
				cfg.APIKey = loginKey
			}

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			cfgPath, _ := config.FilePath()
			ui.SuccessPrint("Saved credentials to %s", cfgPath)
			return nil
		},
	}
	loginCmd.Flags().StringVar(&loginToken, "token", "", "Session token or Bearer JWT")
	loginCmd.Flags().StringVar(&loginKey, "api-key", "", "ElevenLabs API key")
	cmd.AddCommand(loginCmd)

	// Logout
	cmd.AddCommand(&cobra.Command{
		Use:   "logout",
		Short: "Clear stored credentials from config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Clear(); err != nil {
				return err
			}
			ui.SuccessPrint("Stored credentials removed.")
			return nil
		},
	})

	// Import
	cmd.AddCommand(&cobra.Command{
		Use:   "import",
		Short: "Automatically detect and import active session from local Brave / Chrome browser profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Scanning local browser profiles for active ElevenReader sessions...")
			tokens, err := auth.DetectBrowserTokens()
			if err != nil {
				return fmt.Errorf("scanning browser data: %w", err)
			}

			if len(tokens) == 0 {
				return fmt.Errorf("no ElevenReader tokens found in browser storage. Make sure you are logged into https://elevenreader.io in Brave or Chrome")
			}

			fmt.Printf("Found %d candidate tokens. Verifying with API...\n", len(tokens))
			baseURL := flagBaseURL
			if baseURL == "" {
				baseURL = config.DefaultBaseURL
			}

			var workingToken string
			for _, t := range tokens {
				if auth.VerifyToken(baseURL, t) {
					workingToken = t
					break
				}
			}

			if workingToken == "" {
				return fmt.Errorf("found candidate tokens, but none were active. Please refresh https://elevenreader.io in your browser and try again")
			}

			cfg, _ := config.Load()
			if cfg == nil {
				cfg = &config.Config{}
			}
			cfg.Token = workingToken
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			cfgPath, _ := config.FilePath()
			ui.SuccessPrint("Successfully imported active session token to %s!", cfgPath)
			fmt.Printf("Token: %s\n", auth.MaskToken(workingToken))
			return nil
		},
	})

	return cmd
}
