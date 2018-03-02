// Package main contains the main nfpm cli source code.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/a8m/envsubst"
	"github.com/alecthomas/kingpin"
	"github.com/gobuffalo/packr"
	"github.com/goreleaser/nfpm"
	_ "github.com/goreleaser/nfpm/deb"
	_ "github.com/goreleaser/nfpm/rpm"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v1"
)

var (
	version = "master"

	app    = kingpin.New("nfpm", "not-fpm packages apps in some formats")
	config = app.Flag("config", "config file").
		Default("nfpm.yaml").
		Short('f').
		String()

	pkgCmd = app.Command("pkg", "package based on the config file").Alias("package")
	target = pkgCmd.Flag("target", "where to save the generated package").
		Default("/tmp/foo.deb").
		Short('t').
		String()

	initCmd = app.Command("init", "create an empty config file")
)

func main() {
	app.Version(version)
	app.VersionFlag.Short('v')
	app.HelpFlag.Short('h')
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case initCmd.FullCommand():
		if err := initFile(*config); err != nil {
			kingpin.Fatalf(err.Error())
		}
		fmt.Printf("created config file from example: %s\n", *config)
	case pkgCmd.FullCommand():
		if err := doPackage(*config, *target); err != nil {
			kingpin.Fatalf(err.Error())
		}
		fmt.Printf("created package: %s\n", *target)
	}
}

func initFile(config string) error {
	box := packr.NewBox("./nfpm.yaml.example")
	return ioutil.WriteFile(config, box.Bytes("."), 0666)
}

func doPackage(config, target string) error {
	format := filepath.Ext(target)[1:]
	bts, err := ioutil.ReadFile(config)
	if err != nil {
		return errors.Wrapf(err, "error read config file '%s'", target)
	}
	bts, err = envsubst.Bytes(bts)
	if err != nil {
		return errors.Wrap(err, "error substituting environment variables")
	}
	fmt.Println(string(bts))

	var info nfpm.Info
	err = yaml.Unmarshal(bts, &info)
	if err != nil {
		return errors.Wrap(err, "error parsing yml configuration")
	}
	fmt.Printf("using %s packager...\n", format)
	pkg, err := nfpm.Get(format)
	if err != nil {
		return errors.Wrapf(err, "cannot use packager '%s'", format)
	}

	f, err := os.Create(target)
	if err != nil {
		return errors.Wrapf(err, "cannot create target file '%s'", target)
	}
	err = pkg.Package(nfpm.WithDefaults(info), f)
	return errors.Wrapf(err, "packager '%s' failed", format)
}
