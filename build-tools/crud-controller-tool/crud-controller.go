package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

var model string
var modelPath string
var name string
var base64imgFlag string
var mediaPrefix string
var dateMediaPrefix bool
var packageName string
var saveTo string

//go:embed "crud-controller.gotpl"
var crudController []byte

func init() {
	flag.StringVar(&model, "model", "", "model name Users, Post, Comment etc.")
	flag.StringVar(&packageName, "package", "controllers", "package name (default controllers)")
	flag.StringVar(&name, "name", "", "crud name user,post,comment etc")
	flag.StringVar(&base64imgFlag, "base64img", "", "field with base64 images field ';' separate (Image;AddImg)")
	flag.StringVar(&mediaPrefix, "prefix", "", "add prefix to media paths (/sources/)")
	flag.StringVar(&modelPath, "path", "", "add path to model (example axgrid.com/internal/data/source)")
	flag.StringVar(&saveTo, "to", "./target/generated-sources/controllers/", "save file path (./target/generated-sources/transport/)")
	flag.BoolVar(&dateMediaPrefix, "datePrefix", true, "add date prefix to media paths (/2011/01/02/)")

	flag.Parse()
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05,000"}).Level(zerolog.DebugLevel)

	getThumbnailOptions(base64imgFlag)

	if model == "" {
		log.Panic().Msg("model not set, user --model={name}")
	}
	model = strings.Title(model)
	if name == "" {
		name = strings.ToLower(model)
	}
	log.Info().Str("model", model).Str("name", name).Msg("ok")

	var base64Fields []string
	for _, s := range strings.Split(base64imgFlag, ",") {
		base64Fields = append(base64Fields, s)
	}

	type TemplateOpts struct {
		Package           string
		Model             string
		MethodName        string
		Name              string
		Base64Fields      []BaseImg64Field
		MediaPrefix       string
		DateMediaPrefix   bool
		ImportModelPath   string
		ImportModelPrefix string
	}

	opts := TemplateOpts{
		Package:         packageName,
		Model:           model,
		Name:            name,
		Base64Fields:    getThumbnailOptions(base64imgFlag),
		MediaPrefix:     mediaPrefix,
		DateMediaPrefix: dateMediaPrefix,
		MethodName:      fmt.Sprintf("New%sController", model),
	}

	if modelPath != "" {
		opts.ImportModelPath = modelPath
		opts.Model = "mp." + opts.Model
		//opts.ImportModelPrefix = strings.Split(modelPath, "/")[len(strings.Split(modelPath, "/"))-1] + "."
	}
	if saveTo == "" {
		println(render("crud", opts))
	} else {
		fp, err := filepath.Abs(saveTo)
		if err != nil {
			log.Fatal().Err(err).Msgf("abs")
		}
		createDir(fp)
		err = os.WriteFile(path.Join(fp, name+".go"), []byte(render("crud", opts)), 0644)
		if err != nil {
			log.Fatal().Err(err).Msgf("write file")
		}
	}
}

type ThumbOpts struct {
	Width  int
	Height int
}

type BaseImg64Field struct {
	Name  string
	Thumb []ThumbOpts
}

func getThumbnailOptions(fields string) (res []BaseImg64Field) {
	if fields == "" {
		return
	}
	log.Info().Msgf("Thumb:%s", fields)

	pattern := regexp.MustCompile("^(\\w+)(\\[(\\(\\d+,\\d+\\),?)+\\])?$")
	r := regexp.MustCompile(`\(\d+,\d+\)`)
	sr := regexp.MustCompile(`^\((\d+),(\d+)\)$`)

	for _, f := range strings.Split(fields, ";") {
		args := pattern.FindStringSubmatch(f)
		//log.Info().Str("name", args[1]).Str("elements", args[2]).Interface("args", args).Msg("---")
		field := BaseImg64Field{
			Name:  args[1],
			Thumb: []ThumbOpts{},
		}
		if args[2] != "" {
			sm := r.FindAllString(args[2], -1)
			//log.Info().Interface("sm", sm).Msgf("from:%s", args[2])
			for _, el := range sm {
				numArgs := sr.FindStringSubmatch(el)
				//log.Info().Msgf(" - %s:%s", numArgs[1], numArgs[2])
				field.Thumb = append(field.Thumb, ThumbOpts{
					Width:  atoi(numArgs[1]),
					Height: atoi(numArgs[2]),
				})
			}
		}
		res = append(res, field)
	}
	//log.Info().Interface("fields", res).Msg("---")
	return
}

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal().Err(err).Msg("atoi")
	}
	return i
}

func render(name string, data interface{}) string {
	var templateString string
	switch name {
	default:
	case "crud":
		templateString = string(crudController)
	}
	t, err := template.New(name).Funcs(template.FuncMap{}).Parse(templateString)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to create %s template", name)
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, data)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to render %s template", name)
	}
	return tpl.String()
}

func createDir(absDir string) {
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		if err := os.MkdirAll(absDir, 0755); err != nil {
			log.Fatal().Err(err).Msgf("failed to create folder %v", absDir)
		}
	} else if err != nil {
		log.Fatal().Err(err).Msgf("failed to os.Stat %v", absDir)
	}
}
