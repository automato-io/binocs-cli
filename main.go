package main

import (
	"fmt"

	"github.com/automato-io/binocs-cli/cmd"
	"github.com/automato-io/s3update"
)

func main() {
	err := s3update.AutoUpdate(s3update.Updater{
		CurrentVersion: cmd.BinocsVersion,
		S3VersionKey:   "VERSION",
		S3Bucket:       "binocs-download-website",
		S3ReleaseKey:   "binocs_{{VERSION}}_{{OS}}_{{ARCH}}.tgz",
		ChecksumKey:    "binocs_{{VERSION}}_{{OS}}_{{ARCH}}_checksum.txt",
		Verbose:        true,
	})
	if err != nil {
		fmt.Println(err)
	}
	cmd.Execute()
}
