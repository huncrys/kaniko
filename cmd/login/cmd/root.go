/*
Copyright 2018 Google LLC

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
	"io"
	"os"
	"strings"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/types"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type loginOptions struct {
	serverAddress string
	user          string
	password      string
	passwordStdin bool
}

var opts loginOptions

func init() {
	flags := RootCmd.Flags()

	flags.StringVarP(&opts.user, "username", "u", "", "Username")
	flags.StringVarP(&opts.password, "password", "p", "", "Password")
	flags.BoolVarP(&opts.passwordStdin, "password-stdin", "", false, "Take the password from stdin")
}

var RootCmd = &cobra.Command{
	Use:                   "login [OPTIONS] [SERVER]",
	DisableFlagsInUseLine: true,
	Short:                 "Log in to a registry",
	Long:                  "Authenticate to a registry.\nDefaults to Docker Hub if no server is specified.",
	Args:                  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		server := name.DefaultRegistry
		if len(args) > 0 {
			server = args[0]
		}

		registry, err := name.NewRegistry(server)
		if err != nil {
			return err
		}

		opts.serverAddress = registry.Name()

		return login(opts)
	},
}

func login(opts loginOptions) error {
	if opts.password != "" && opts.passwordStdin {
		return errors.New("--password and --password-stdin are mutually exclusive")
	}
	if opts.passwordStdin {
		if opts.user == "" {
			return errors.New("Must provide --username with --password-stdin")
		}

		contents, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		opts.password = strings.TrimSuffix(string(contents), "\n")
		opts.password = strings.TrimSuffix(opts.password, "\r")
	}

	if opts.user == "" && opts.password == "" {
		return errors.New("username and password required")
	}

	cf, err := config.Load(os.Getenv("DOCKER_CONFIG"))
	if err != nil {
		return err
	}

	creds := cf.GetCredentialsStore(opts.serverAddress)
	if opts.serverAddress == name.DefaultRegistry {
		opts.serverAddress = authn.DefaultAuthKey
	}

	if err := creds.Store(types.AuthConfig{
		ServerAddress: opts.serverAddress,
		Username:      opts.user,
		Password:      opts.password,
	}); err != nil {
		return err
	}

	if err := cf.Save(); err != nil {
		return err
	}

	logrus.Printf("Credentials stored for '%s' in '%s'", opts.serverAddress, cf.Filename)
	return nil
}
