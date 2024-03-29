package project

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v2"

	"github.com/sado0823/go-kitx/cmd/kitx/internal"
)

var (
	defaultTpl     string
	timeoutSeconds int64
)

func init() {
	defaultTpl = "https://github.com/sado0823/go-kitx-tpl.git"
	timeoutSeconds = 60
}

var (
	flagName  = "name"
	flagNomod = "nomod"
)

var repoAddIgnores = []string{
	".git", ".github", "api", "README.md", "LICENSE", "go.mod", "go.sum", "third_party", "openapi.yaml", ".gitignore",
}

type project struct {
	name string
	path string
}

func Cmd() *cli.Command {
	return &cli.Command{
		Name:    "new",
		Aliases: []string{"n"},
		Usage:   "generate project from template",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagName,
				Usage: "set project name",
			},
			&cli.BoolFlag{
				Name:  flagNomod,
				Usage: "new project with no go mod file",
			},
		},
		Action: func(cCtx *cli.Context) error {
			fmt.Println("generate project from template")

			pwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			fmt.Println("getwd", pwd)

			projectName := cCtx.String(flagName)
			projectNomod := cCtx.Bool(flagNomod)
			fmt.Println(projectName, projectNomod, "......")
			if len(projectName) == 0 {
				confirm := &survey.Input{
					Message: "What's your project name❓",
					Help:    "please input a project name",
				}
				if err := survey.AskOne(confirm, &projectName); err != nil {
					return err
				}
				if len(projectName) == 0 {
					return fmt.Errorf("invalid project name:%s", projectName)
				}
			}

			pj := &project{name: path.Base(projectName), path: projectName}

			fmt.Printf("project:%#v \n", pj)

			timeoutCtx, cancelFunc := context.WithTimeout(context.Background(), time.Second*time.Duration(timeoutSeconds))
			defer cancelFunc()

			if projectNomod {
				if _, e := os.Stat(path.Join(pwd, "go.mod")); os.IsNotExist(e) {
					return fmt.Errorf("🚫 go.mod don't exists in %s", pwd)
				}

				mod, e := internal.GoModulePath(path.Join(pwd, "go.mod"))
				if e != nil {
					return e
				}

				err = pj.add(timeoutCtx, pwd, defaultTpl, "master", mod)
			} else {
				err = pj.new(timeoutCtx, pwd, defaultTpl, "master")
			}
			return err
		},
	}
}

func (p *project) overwrite(ctx context.Context, toDir string) error {
	pDir := path.Join(toDir, p.name)
	if _, err := os.Stat(pDir); !os.IsNotExist(err) {
		fmt.Printf("💢 %s already exists\n", p.name)
		override := false
		prompt := &survey.Confirm{
			Message: "📂 Do you want to override the folder ?",
			Help:    "Delete the existing folder and create the project.",
		}
		e := survey.AskOne(prompt, &override)
		if e != nil {
			return e
		}
		if !override {
			return err
		}
		_ = os.RemoveAll(pDir)
	}
	return nil
}

func (p *project) gitClone(ctx context.Context, toDir, tplUrl, tplBranch string) error {
	var (
		errGit error
		errOs  error
	)
	if _, errOs = os.Stat(toDir); !os.IsNotExist(errOs) {
		errGit = internal.GitPull(ctx, tplUrl, toDir)
	} else {
		errGit = internal.GitClone(ctx, tplUrl, tplBranch, toDir)
	}
	if errGit != nil {
		return fmt.Errorf("error git:%+v, os stat err:%+v", errGit, errOs)
	}

	return nil
}

func (p *project) tplHome(ctx context.Context, tplBranch string) string {
	home := internal.KitxHomeWithDir("repo/")
	return path.Join(home, "kitx-tpl@"+tplBranch)
}

func (p *project) add(ctx context.Context, toDir, tplUrl, tplBranch, modPath string) error {
	if err := p.overwrite(ctx, toDir); err != nil {
		return err
	}

	var (
		pDir    = path.Join(toDir, p.path)
		homeTPL = p.tplHome(ctx, tplBranch)
	)

	fmt.Printf("💕 adding project: %s ...\n", pDir)
	fmt.Println("homeTPL: ", homeTPL)

	if err := p.gitClone(ctx, homeTPL, tplUrl, tplBranch); err != nil {
		return err
	}

	// go mod file
	tplMod, err := internal.GoModulePath(path.Join(homeTPL, "go.mod"))
	if err != nil {
		return err
	}

	pwdMod, err := internal.GoModulePath(path.Join(toDir, "go.mod"))
	if err != nil {
		return err
	}

	pwdMod = path.Join(pwdMod, p.path)
	toReplace := []string{tplMod, pwdMod, path.Join(p.path, "api"), "api"}

	return cpDir(homeTPL, pDir, toReplace, repoAddIgnores)
}

func (p *project) new(ctx context.Context, toDir string, tplUrl, tplBranch string) error {
	if err := p.overwrite(ctx, toDir); err != nil {
		return err
	}

	pDir := path.Join(toDir, p.path)
	fmt.Printf("💚 creating project: %s ...\n", pDir)

	homeTPL := p.tplHome(ctx, tplBranch)
	fmt.Println("homeTPL: ", homeTPL)

	if err := p.gitClone(ctx, homeTPL, tplUrl, tplBranch); err != nil {
		return err
	}

	// go mod file
	modPath, err := internal.GoModulePath(path.Join(homeTPL, "go.mod"))
	if err != nil {
		return err
	}

	fmt.Println("modPath", modPath)
	fmt.Println("p.path", p.path)
	fmt.Println("homeTPL", homeTPL)

	err = cpDir(homeTPL, pDir, []string{modPath, p.path}, []string{".git", ".github"})
	if err != nil {
		return err
	}

	fmt.Printf("%s has been created !!! \n", p.name)
	fmt.Println("$ cd ", p.name)
	fmt.Println("$ go generate ./...")
	fmt.Println("$ cd cmd && go run main.go")

	return nil
}

func cpDir(src, dst string, toReplace, needIgnore []string) error {
	srcFileInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcFileInfo.Mode()); err != nil {
		return err
	}

	srcDirEntries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range srcDirEntries {
		if inArr(entry.Name(), needIgnore) {
			continue
		}
		srcFilePath := path.Join(src, entry.Name())
		dstFilePath := path.Join(dst, entry.Name())
		var e error
		if entry.IsDir() {
			e = cpDir(srcFilePath, dstFilePath, toReplace, needIgnore)
		} else {
			e = cpFile(srcFilePath, dstFilePath, toReplace)
		}
		if e != nil {
			return e
		}
	}
	return nil
}

func cpFile(src, dst string, toReplace []string) error {
	srcStat, err := os.Stat(src)
	if err != nil {
		return err
	}
	srcFile, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	var old string
	for i, replace := range toReplace {
		if i%2 == 0 {
			old = replace
			continue
		}
		srcFile = bytes.ReplaceAll(srcFile, []byte(old), []byte(replace))
	}
	return os.WriteFile(dst, srcFile, srcStat.Mode())
}

func inArr(key string, sets []string) bool {
	for _, set := range sets {
		if key == set {
			return true
		}
	}

	return false
}
