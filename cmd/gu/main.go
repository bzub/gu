package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gu-io/gu/generators"
	"github.com/influx6/faux/fmtwriter"
	"github.com/influx6/faux/metrics"
	"github.com/influx6/faux/metrics/custom"
	"github.com/influx6/moz/ast"
	"github.com/influx6/moz/gen"
	"github.com/influx6/moz/gen/templates"
	cli "gopkg.in/urfave/cli.v2"
)

var (
	version     = "0.0.1"
	defaultName = "manifests"
	commands    = []*cli.Command{}

	gupath = "github.com/gu-io/gu"
)

func main() {
	initCommands()

	app := &cli.App{}
	app.Name = "Gu"
	app.Version = version
	app.Commands = commands
	app.Usage = `Gu CLI tooling to make developing UI projects easier.`

	app.Run(os.Args)
}

func capitalize(val string) string {
	return strings.ToUpper(val[:1]) + val[1:]
}

var badSymbols = regexp.MustCompile(`[(|\-|_|\W|\d)+]`)
var notAllowed = regexp.MustCompile(`[^(_|\w|\d)+]`)
var descore = regexp.MustCompile("-")

func validateName(val string) bool {
	return notAllowed.MatchString(val)
}

func initCommands() {

	commands = append(commands, &cli.Command{
		Name:        "component",
		Usage:       "gu component <component-name>",
		Description: `Generates a new boilerplate for component package.`,
		Flags:       []cli.Flag{},
		Action: func(ctx *cli.Context) error {
			args := ctx.Args()

			if args.Len() == 0 {
				return errors.New("Please provide the name for your package")
			}

			component := args.First()
			currentDir, err := os.Getwd()
			if err != nil {
				return err
			}

			directives, err := generators.ComponentPackageGenerator(ast.AnnotationDeclaration{Arguments: []string{component}}, ast.PackageDeclaration{FilePath: currentDir}, ast.Package{})
			if err != nil {
				return err
			}

			for _, directive := range directives {
				if directive.Dir != "" {
					coDir := filepath.Join(currentDir, directive.Dir)

					if _, err := os.Stat(coDir); err != nil {
						fmt.Printf("- Creating package directory: %q\n", coDir)
						if err := os.MkdirAll(coDir, 0700); err != nil && err != os.ErrExist {
							return err
						}
					}

				}

				if directive.Writer == nil {
					continue
				}

				coFile := filepath.Join(currentDir, directive.Dir, directive.FileName)

				if _, err := os.Stat(coFile); err == nil {
					if directive.DontOverride {
						continue
					}
				}

				dFile, err := os.Create(coFile)
				if err != nil {
					return err
				}

				if _, err := directive.Writer.WriteTo(dFile); err != nil {
					return err
				}

				rel, _ := filepath.Rel(currentDir, coFile)
				fmt.Printf("- Add file to package directory: %q\n", rel)

				dFile.Close()
			}

			return nil
		},
	})

	commands = append(commands, &cli.Command{
		Name:        "driver",
		Usage:       "gu driver <driver-name>",
		Description: `Generates a new boilerplate for app driver package which launches the package in the system desired. .e.g js for gopherjs`,
		Flags:       []cli.Flag{},
		Action: func(ctx *cli.Context) error {
			args := ctx.Args()

			if args.Len() == 0 {
				return errors.New("Please provide the name for your package")
			}

			driver := args.First()
			currentDir, err := os.Getwd()
			if err != nil {
				return err
			}

			var directives []gen.WriteDirective

			switch driver {
			case "js":
				directives, err = generators.JSDriverGenerator(ast.AnnotationDeclaration{}, ast.PackageDeclaration{FilePath: currentDir}, ast.Package{})
				break
			default:
				return fmt.Errorf("Driver %s not supported yet", driver)
			}

			if err != nil {
				return err
			}

			// appDir := filepath.Join(currentDir, component)

			for _, directive := range directives {
				if directive.Dir != "" {
					coDir := filepath.Join(currentDir, directive.Dir)

					if _, err := os.Stat(coDir); err != nil {
						drel, _ := filepath.Rel(currentDir, coDir)
						fmt.Printf("- Creating package directory: %q\n", drel)

						if err := os.MkdirAll(coDir, 0700); err != nil && err != os.ErrExist {
							return err
						}
					}

				}

				if directive.Writer == nil {
					continue
				}

				coFile := filepath.Join(currentDir, directive.Dir, directive.FileName)

				if _, err := os.Stat(coFile); err == nil {
					if directive.DontOverride {
						continue
					}
				}

				dFile, err := os.Create(coFile)
				if err != nil {
					return err
				}

				if _, err := directive.Writer.WriteTo(dFile); err != nil {
					return err
				}

				rel, _ := filepath.Rel(currentDir, coFile)
				fmt.Printf("- Add file to package directory: %q\n", rel)

				dFile.Close()
			}

			return nil
		},
	})

	commands = append(commands, &cli.Command{
		Name:        "app",
		Usage:       "gu app <package-name>",
		Description: `Generates a new boilerplate for app package.`,
		Flags:       []cli.Flag{},
		Action: func(ctx *cli.Context) error {
			args := ctx.Args()

			if args.Len() == 0 {
				return errors.New("Please provide the name for your package")
			}

			component := args.First()
			currentDir, err := os.Getwd()
			if err != nil {
				return err
			}

			directives, err := generators.GuPackageGenerator(ast.AnnotationDeclaration{Arguments: []string{component}}, ast.PackageDeclaration{FilePath: currentDir}, ast.Package{})
			if err != nil {
				return err
			}

			// appDir := filepath.Join(currentDir, component)

			for _, directive := range directives {
				if directive.Dir != "" {
					coDir := filepath.Join(currentDir, directive.Dir)

					if _, err := os.Stat(coDir); err != nil {
						drel, _ := filepath.Rel(currentDir, coDir)
						fmt.Printf("- Creating package directory: %q\n", drel)

						if err := os.MkdirAll(coDir, 0700); err != nil && err != os.ErrExist {
							return err
						}
					}

				}

				if directive.Writer == nil {
					// fmt.Printf("-- [NoWriter]Skipping operation in package directory: %q\n", directive.Dir)
					continue
				}

				coFile := filepath.Join(currentDir, directive.Dir, directive.FileName)

				if _, err := os.Stat(coFile); err == nil {
					if directive.DontOverride {
						continue
					}
				}

				dFile, err := os.Create(coFile)
				if err != nil {
					return err
				}

				if _, err := directive.Writer.WriteTo(dFile); err != nil {
					return err
				}

				rel, _ := filepath.Rel(currentDir, coFile)
				fmt.Printf("- Add file to package directory: %q\n", rel)

				dFile.Close()
			}

			return nil
		},
	})

	commands = append(commands, &cli.Command{
		Name:        "generate",
		Usage:       "gu generate",
		Description: "Generate will call needed code generators to create project assets and files as declared by the project and it's sources",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "inputdir",
				Aliases: []string{"dir"},
				Usage:   "dir=./my-gu-project",
			},
		},
		Action: func(ctx *cli.Context) error {
			indir := ctx.String("inputdir")

			if indir == "" {
				cdir, err := os.Getwd()
				if err != nil {
					return err
				}

				indir = cdir
			}

			register := ast.NewAnnotationRegistry()

			generators.RegisterGenerators(register)

			// Register @assets annotation for our registery as well.
			register.Register("assets", AssetsAnnotationGenerator)

			events := metrics.New(custom.BlockDisplay(os.Stdout))
			pkg, err := ast.ParseAnnotations(events, indir)
			if err != nil {
				events.Emit(metrics.Error(err), metrics.With("dir", indir),
					metrics.Message("Failed to parse package annotations"))
				return err
			}

			if err := ast.Parse(indir, events, register, false, pkg...); err != nil {
				events.Emit(metrics.Error(err), metrics.With("dir", indir),
					metrics.Message("Failed to parse package annotations"))
				return err
			}

			return nil
		},
	})

}

// FindLowerByStat searches the path line down until it's roots to find the directory with the giving
// dirName matching else returns an error.
func findLowerByStat(root string, path string, dirName string, dirOnly bool) (string, error) {
	path = filepath.Clean(path)

	if path == "." {
		return "", errors.New("'" + dirName + "' path not found")
	}

	// Let's attempt to see if there is a dirName in this path and if it's a
	// directory.
	possiblePath := filepath.Join(root, path, dirName)
	possibleStat, err := os.Stat(possiblePath)
	if err == nil {
		if dirOnly && !possibleStat.IsDir() {
			return findLower(filepath.Join(path, ".."), dirName)
		}

		return filepath.Join(path, dirName), nil
	}

	return findLowerByStat(root, filepath.Join(path, ".."), dirName, dirOnly)
}

// Searches the path line down until it's roots to find the directory with the giving
// dirName matching else returns an error.
func findLower(path string, dirName string) (string, error) {
	path = filepath.Clean(path)

	if path == "." {
		return "", errors.New("'" + dirName + "' path not found")
	}

	if filepath.Base(path) == dirName {
		return path, nil
	}

	return findLower(filepath.Join(path, ".."), dirName)
}

func writeFile(targetFile string, data []byte) error {
	file, err := os.Create(targetFile)
	if err != nil {
		return err
	}

	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return err
	}

	return nil
}

// AssetsAnnotationGenerator defines a package level annotation generator which builds a go package in
// root of the package it appears in to provide a means to quickly draft all file contents into the created
// package.
// Annotation: @assets
// Arguments(Optional): (PackageName, FileExtensionsToSupport, DirectorNameForFiles)
// 	e.g @assets(assets, ".tml : .bol : .go : .js", mytemplates).
func AssetsAnnotationGenerator(toDir string, an ast.AnnotationDeclaration, pkgDeclr ast.PackageDeclaration, pkg ast.Package) ([]gen.WriteDirective, error) {
	var directives []gen.WriteDirective

	var extensions []string

	pkgName := "assets"
	contentFileName := "files"

	if argLen := len(an.Arguments); argLen != 0 {
		pkgName = an.Arguments[0]

		if argLen > 1 {
			extensions = strings.Split(an.Arguments[1], ":")
		}

		if argLen > 2 {
			contentFileName = an.Arguments[2]
		}
	}

	genFile := gen.Package(
		gen.Name(pkgName),
		gen.Text("//go:generate go run generate.go"),
	)

	directives = append(directives, gen.WriteDirective{
		Writer:   fmtwriter.New(genFile, true, true),
		FileName: fmt.Sprintf("%s.go", pkgName),
		Dir:      pkgName,
	})

	directives = append(directives, gen.WriteDirective{
		Dir: filepath.Join(pkgName, contentFileName),
	})

	mainFile := gen.Block(
		gen.Commentary(
			gen.Text("+build ignore"),
		),
		gen.Text("\n"),
		gen.Text("\n"),
		gen.Package(
			gen.Name("main"),
			gen.Imports(
				gen.Import("fmt", ""),
				gen.Import("path/filepath", ""),
				gen.Import("github.com/influx6/moz/gen", ""),
				gen.Import("github.com/influx6/moz/utils", ""),
				gen.Import("github.com/influx6/faux/vfiles", ""),
				gen.Import("github.com/influx6/faux/fmtwriter", ""),
				gen.Import("github.com/influx6/faux/metrics", ""),
				gen.Import("github.com/influx6/faux/metrics/custom", ""),
			),
			gen.Function(
				gen.Name("main"),
				gen.Constructor(),
				gen.Returns(),
				gen.Block(
					gen.SourceText(
						string(templates.Must("assets/assets.tml")),
						struct {
							Extensions       []string
							TargetDir        string
							Package          string
							GenerateTemplate string
						}{
							TargetDir:  contentFileName,
							Extensions: extensions,
							Package:    pkgName,
							GenerateTemplate: `{{range $key, $value := .Files}}files[{{quote $key}}] = []byte("{{$value}}")
							{{end}}`,
						},
					),
				),
			),
		),
	)

	directives = append(directives, gen.WriteDirective{
		Writer:   fmtwriter.New(mainFile, true, true),
		FileName: "generate.go",
		Dir:      pkgName,
	})

	return directives, nil
}
