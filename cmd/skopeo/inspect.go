package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/containers/image/docker"
	"github.com/containers/image/manifest"
	"github.com/containers/image/transports"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// inspectOutput is the output format of (skopeo inspect), primarily so that we can format it with a simple json.MarshalIndent.
type inspectOutput struct {
	Name          string `json:",omitempty"`
	Tag           string `json:",omitempty"`
	Digest        digest.Digest
	RepoTags      []string
	Created       time.Time
	DockerVersion string
	Labels        map[string]string
	Architecture  string
	Os            string
	Layers        []string
}

var inspectCmd = cli.Command{
	Name:  "inspect",
	Usage: "Inspect image IMAGE-NAME",
	Description: fmt.Sprintf(`
	Return low-level information about "IMAGE-NAME" in a registry/transport

	Supported transports:
	%s

	See skopeo(1) section "IMAGE NAMES" for the expected format
	`, strings.Join(transports.ListNames(), ", ")),
	ArgsUsage: "IMAGE-NAME",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "cert-dir",
			Value: "",
			Usage: "use certificates at `PATH` (*.crt, *.cert, *.key) to connect to the registry",
		},
		cli.BoolTFlag{
			Name:  "tls-verify",
			Usage: "require HTTPS and verify certificates when talking to container registries (defaults to true)",
		},
		cli.BoolFlag{
			Name:  "raw",
			Usage: "output raw manifest",
		},
		cli.StringFlag{
			Name:  "creds",
			Value: "",
			Usage: "Use `USERNAME[:PASSWORD]` for accessing the registry",
		},
	},
	Action: func(c *cli.Context) (retErr error) {
		img, err := parseImage(c)
		if err != nil {
			return err
		}

		defer func() {
			if err := img.Close(); err != nil {
				retErr = errors.Wrapf(retErr, fmt.Sprintf("(could not close image: %v) ", err))
			}
		}()

		rawManifest, _, err := img.Manifest()
		if err != nil {
			return err
		}
		if c.Bool("raw") {
			_, err := c.App.Writer.Write(rawManifest)
			if err != nil {
				return fmt.Errorf("Error writing manifest to standard output: %v", err)
			}
			return nil
		}
		imgInspect, err := img.Inspect()
		if err != nil {
			return err
		}
		outputData := inspectOutput{
			Name: "", // Possibly overridden for a docker.Image.
			Tag:  imgInspect.Tag,
			// Digest is set below.
			RepoTags:      []string{}, // Possibly overriden for a docker.Image.
			Created:       imgInspect.Created,
			DockerVersion: imgInspect.DockerVersion,
			Labels:        imgInspect.Labels,
			Architecture:  imgInspect.Architecture,
			Os:            imgInspect.Os,
			Layers:        imgInspect.Layers,
		}
		outputData.Digest, err = manifest.Digest(rawManifest)
		if err != nil {
			return fmt.Errorf("Error computing manifest digest: %v", err)
		}
		if dockerImg, ok := img.(*docker.Image); ok {
			outputData.Name = dockerImg.SourceRefFullName()
			outputData.RepoTags, err = dockerImg.GetRepositoryTags()
			if err != nil {
				// some registries may decide to block the "list all tags" endpoint
				// gracefully allow the inspect to continue in this case. Currently
				// the IBM Bluemix container registry has this restriction.
				if !strings.Contains(err.Error(), "401") {
					return fmt.Errorf("Error determining repository tags: %v", err)
				}
				logrus.Warnf("Registry disallows tag list retrieval; skipping")
			}
		}
		out, err := json.MarshalIndent(outputData, "", "    ")
		if err != nil {
			return err
		}
		fmt.Fprintln(c.App.Writer, string(out))
		return nil
	},
}
