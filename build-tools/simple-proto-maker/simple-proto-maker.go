package main

import (
	ax_tools "bot_tasker/shared/axtools"
	shell_runner "bot_tasker/shared/shell-runner"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var outRoot = flag.String("out", "./target/generated-sources/proto/", "proto generated source out path")
var root = flag.String("root", "./protobufs", "proto project folder")
var pprotoc = flag.String("protoc", "protoc", "protoc path")
var protoc = "protoc"
var protocArgs = " --go_out=./ "
var cwd string

type ProtoFile struct {
	Module  string
	Project string
	Path    string
}

func main() {
	flag.Parse()
	protoc = *pprotoc
	ax_tools.InitLogger("info")

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("get cwd error")
	}
	_ = os.RemoveAll("tmp")
	if err := os.Mkdir("tmp", 0755); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(*outRoot, 0755); err != nil {
		panic(err)
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	cwd = cwd + "/" + *root

	//var fileList []ProtoFile
	log.Info().Msgf("build protobuf from:%s to:%s", filepath.Clean(cwd), *outRoot)

	var protoMap = map[string][]ProtoFile{}

	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".proto" {
			path = strings.Replace(path, filepath.Clean(cwd)+"/", "", 1)
			log.Info().Msgf("Found file %s", path)

			//protoMap[ax_tools.GetFirstDir(Path)]
			pName := ax_tools.GetFirstDir(path)
			protoMap[pName] = append(protoMap[pName], ProtoFile{
				Path:    path,
				Project: pName,
				Module:  filepath.Base(path),
			})
		}
		return nil
	})

	log.Info().Interface("protoMap", protoMap).Msg("ProtoMap created")

	protocDoneChan := make(chan *shell_runner.TimedCommandResult)
	for p, fileList := range protoMap {
		cmd := fmt.Sprintf("%s %s %s", protoc, protocArgs, createCommand(fileList, p))
		//log.Info().Msg(cmd)
		go shell_runner.RunTimedCommand(cmd, protocDoneChan)
		res := <-protocDoneChan
		if res.Error != nil {
			log.Fatal().Err(res.Error).Msg("Shell error")
			panic(err)
		}
		log.Info().Str("project", p).Msg(res.StdOut)
	}
}

func createCommand(protos []ProtoFile, project string) string {
	files := ""
	opts := ""
	for _, p := range protos {
		opts += fmt.Sprintf("--go_opt=M%s=%s/ ", p.Module, "./"+filepath.Join(*outRoot, project))
		files += fmt.Sprintf("%s/%s ", *root, p.Path)
	}
	return fmt.Sprintf("--proto_path=%s/%s/ ", *root, project) + opts + " " + files
}
