package main

import (
	"bot_tasker/shared/ospathlib"
	shell_runner "bot_tasker/shared/shell-runner"
	"fmt"
	"github.com/emicklei/proto"
	_ "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/logrusorgru/aurora"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type ProtoFile struct {
	FileName        string
	FileDirPath     string
	FileOSPath      string
	ProtoPackage    string
	GoPackage       string
	SourceGoPackage string
	Imports         []string
	Services        []string
	GenerateClient  bool
}

type DirNode struct {
	Name           string
	DirAbsPath     string
	SubDirs        []*DirNode
	Files          []ProtoFile
	NonProjectPart bool
}

var parsedProtoFiles uint32

var ignoredPaths = map[string]bool{
	"node_modules": true,
	"vendor":       true,
}

const (
	protoExtension = ".proto"
)

func main() {
	tms := time.Now()
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	_ = os.RemoveAll("tmp")
	if err := os.Mkdir("tmp", 0755); err != nil {
		panic(err)
	}

	cwd = cwd + "/"

	if err := os.MkdirAll(filepath.Join(cwd, "target/generated-sources"), 0755); err != nil {
		panic(err)
	}
	log.Printf("cwd:%s", cwd)
	sources := &DirNode{Name: "/"}

	err = filepath.Walk(cwd, func(path string, info fs.FileInfo, err error) error {
		return fileWalker(path, info, err, sources)
	})

	log.Printf("Found and parsed %d proto files\n", aurora.Yellow(parsedProtoFiles))
	log.Println("Checking if directory structure is in line with our expectations.")

	// now, let's remove non-project parts from our view
	if len(sources.SubDirs) != 1 {
		var dirList []string

		for _, dir := range sources.SubDirs {
			dirList = append(dirList, dir.Name)
		}

		log.Fatalf(`
We've found too many (or zero) subdirs on default zero level of public path, not sure what happened.
Here is what we see at the very system root of OS we're building on: %v\n
`, dirList)
	}

	sources = findRealRoot(cwd, sources.SubDirs[0])
	wg := sync.WaitGroup{}
	wg.Add(len(sources.SubDirs))

	for _, dir := range sources.SubDirs {
		go func(dir *DirNode) {
			log.Printf("Building %s protobuf source package.", aurora.Green(dir.Name))
			res := buildProtobufPackage(cwd, dir)
			if res.StdOut != "" {
				log.Println(res.StdOut)
			}
			if res.StdErr != "" {
				log.Println(res.StdErr)
			}
			if res.Error != nil {
				log.Fatal("got an error: ", res.Error)
			}
			wg.Done()
		}(dir)
	}

	wg.Wait()

	_ = os.RemoveAll("tmp")

	log.Println("done in ", time.Since(tms))
}

/*
protoc --proto_path=protobufs/whitelabel/general/ --go_out=target/generated-sources/
	--go_opt=Mcommons.proto=wl/commons protobufs/whitelabel/general/commons.proto
*/

func buildProtobufPackage(cwd string, dir *DirNode) *shell_runner.TimedCommandResult {
	files := getAllFiles(dir, "")
	// processing deps
	for _, file := range files {
		currentFileShortPath, err := ospathlib.SubtractPath(file.FileOSPath, cwd)
		log.Printf("processing imports of %s\n",
			aurora.Yellow(currentFileShortPath))
		if err != nil {
			panic(err)
		}
		for _, fImport := range file.Imports {
			// looking over all files to see if that's it
			fImportFullPath := fmt.Sprintf("%s/%s", file.FileDirPath, fImport)
			fImportShortPath, err := ospathlib.SubtractPath(fImportFullPath, cwd)
			if err != nil {
				panic(err)
			}
			// fImportFName := getFileName(fImport)
			for _, f := range files {
				if fImportFullPath == f.FileOSPath {
					log.Printf("%s [%s] imports `%s` [%s]\n",
						aurora.Yellow(file.FileName),
						aurora.Yellow(currentFileShortPath),
						aurora.Green(fImport),
						fImportShortPath)
					break
				}
			}
		}
	}

	var protobufPaths []string
	var goOptParams []string
	var gogoOptParams []string
	var goGrpcOptParams []string
	var servicesPaths []string
	var clientEntryPoints []string

	for _, file := range files {
		rel, err := filepath.Rel(dir.DirAbsPath, file.FileOSPath)
		if err != nil {
			panic(err)
		}

		protobufPaths = append(protobufPaths, file.FileOSPath)
		goOptParams = append(goOptParams, fmt.Sprintf("--go_opt=M%s=%s", rel, strings.TrimPrefix(file.GoPackage, "/")))
		gogoOptParams = append(gogoOptParams, fmt.Sprintf("M%s=%s", rel, strings.TrimPrefix(file.GoPackage, "/")))
		goGrpcOptParams = append(goGrpcOptParams, fmt.Sprintf("--go-grpc_opt=M%s=%s", rel, strings.TrimPrefix(file.GoPackage, "/")))

		if len(file.Services) > 0 {
			servicesPaths = append(servicesPaths, file.FileOSPath)
		}
		if file.GenerateClient {
			clientEntryPoints = append(clientEntryPoints, file.FileOSPath)
		}
	}

	command := fmt.Sprintf("protoc --proto_path=%s --go_out=../ %s %s",
		dir.DirAbsPath,
		strings.Join(goOptParams, " "),
		strings.Join(protobufPaths, " "),
	)
	log.Printf("Running %s\n", aurora.Cyan(command))
	protocDoneChan := make(chan *shell_runner.TimedCommandResult)
	go shell_runner.RunTimedCommand(command, protocDoneChan)
	res := <-protocDoneChan
	if res.Error != nil {
		return res
	}

	for _, gogoEntryPoint := range clientEntryPoints {
		gogoCommand := fmt.Sprintf("protoc --proto_path=%s -I=%s --gogofaster_out=%s:tmp/ %s",
			dir.DirAbsPath,
			dir.DirAbsPath,
			strings.Join(gogoOptParams, ","),
			gogoEntryPoint,
		)
		log.Printf("[Client] Running %s\n", aurora.Cyan(gogoCommand))
		gogoDoneChan := make(chan *shell_runner.TimedCommandResult)
		go shell_runner.RunTimedCommand(gogoCommand, gogoDoneChan)
		gogoRes := <-gogoDoneChan
		res.StdOut += gogoRes.StdOut
		res.StdErr += gogoRes.StdErr
		res.Error = gogoRes.Error
		if gogoRes.Error != nil {
			return res
		}
	}

	for _, servicePath := range servicesPaths {
		servicesDoneChan := make(chan *shell_runner.TimedCommandResult)
		command = fmt.Sprintf("protoc --proto_path=%s --go-grpc_out=../ %s %s",
			dir.DirAbsPath,
			strings.Join(goGrpcOptParams, " "),
			servicePath,
		)
		log.Printf("[GRPC] Running %s\n", aurora.Cyan(command))
		go shell_runner.RunTimedCommand(command, servicesDoneChan)
		acc := <-servicesDoneChan
		res.Error = acc.Error
		res.StdOut = res.StdOut + acc.StdOut
		res.StdErr = res.StdErr + acc.StdErr
		if acc.Error != nil {
			return res
		}
	}

	if res.Error == nil {
		res.Error = postProcessProtoFiles(files)
	}

	if res.Error == nil {
		res.Error = postProcessClientFiles(files)
	}

	return res
}

func (p ProtoFile) outputPath() string {
	fname := p.FileName
	fname = strings.TrimSuffix(fname, protoExtension)
	fname += ".pb.go"
	return fname
}

func postProcessProtoFiles(files []ProtoFile) error {
	for _, file := range files {
		outPath := filepath.Join("..", file.GoPackage, file.outputPath())
		contentBytes, err := os.ReadFile(outPath)
		if err != nil {
			return err
		}
		content := `
// +public !js

` + string(contentBytes)

		if err := os.WriteFile(outPath, []byte(content), 0755); err != nil {
			return err
		}
	}
	return nil
}

func postProcessClientFiles(files []ProtoFile) error {
	for _, file := range files {
		outPath := filepath.Join("tmp", file.GoPackage, file.outputPath())
		realFileName := file.FileName
		realFileName = strings.TrimSuffix(realFileName, protoExtension)
		realFileName = realFileName + ".client.pb.go"
		realOutPath := filepath.Join("..", file.GoPackage, realFileName)
		contentBytes, err := os.ReadFile(outPath)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
		}
		content := `
// +public js

` + string(contentBytes)

		content = strings.ReplaceAll(content, "github.com/gogo/protobuf/proto", "octopus/protobufs")
		if err := os.WriteFile(realOutPath, []byte(content), 0755); err != nil {
			return err
		}

		if err := os.Remove(outPath); err != nil {
			return err
		}
	}

	return nil
}

func getAllFiles(dir *DirNode, rootPackage string) []ProtoFile {
	var subRootPackage string
	if rootPackage == "" {
		subRootPackage = dir.Name
	} else {
		subRootPackage = fmt.Sprintf("%s/%s", rootPackage, dir.Name)
	}

	var result []ProtoFile
	for _, file := range dir.Files {
		if file.SourceGoPackage != "" {
			file.GoPackage = file.SourceGoPackage
		} else if file.ProtoPackage == "" {
			file.GoPackage = filepath.Join("octopus/target/generated-sources/", subRootPackage)
		} else {
			file.GoPackage = filepath.Join("octopus/target/generated-sources/", rootPackage, dir.Name, file.ProtoPackage)
		}
		result = append(result, file)
	}

	for _, dir := range dir.SubDirs {
		for _, dir2 := range getAllFiles(dir, subRootPackage) {
			result = append(result, dir2)
		}
	}

	return result
}

func findRealRoot(cwd string, root *DirNode) *DirNode {
	pathTokens := strings.Split(cwd, "/")
	currentRoot := root
	for _, pathToken := range pathTokens {
		if pathToken == "" {
			continue
		}
		if currentRoot.Name == pathToken {
			// we still see the it...
			currentRoot.NonProjectPart = true
			if len(currentRoot.SubDirs) == 1 && len(currentRoot.Files) == 0 {
				currentRoot = currentRoot.SubDirs[0] // assuming it's the only one...
			} else {
				log.Fatalf(`
Our current working directory is %s.
Now we're attempting to find root for proto files in our project and looking at %s
It's still the part of current working directory because %s == %s.
But we see more then one sub-dir here: %v
Or more then one proto file here: %v
Please start proto-maker from correct working directory (i.e. project root).
`,
					cwd,
					currentRoot.Name,
					currentRoot.Name, pathToken,
					currentRoot.SubDirs, currentRoot.Files)
			}
		}
	}

	return currentRoot
}

func fileWalker(path string, info fs.FileInfo, err error, root *DirNode) error {
	if strings.Contains(path, "cpp-legacy") || info.IsDir() || !strings.HasSuffix(path, protoExtension) {
		return nil
	}
	pathParts := strings.Split(path, string(filepath.Separator))
	for _, part := range pathParts {
		if ignoredPaths[part] {
			return nil
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	parsed, err := proto.NewParser(file).Parse()
	if err != nil {
		return err
	}

	atomic.AddUint32(&parsedProtoFiles, 1)

	protoPackage := ""
	proto.Walk(parsed, proto.WithPackage(func(p *proto.Package) {
		protoPackage = p.Name
	}))

	var imports []string
	proto.Walk(parsed, proto.WithImport(func(i *proto.Import) {
		imports = append(imports, i.Filename)
	}))

	var services []string
	proto.Walk(parsed, proto.WithService(func(i *proto.Service) {
		services = append(services, i.Name)
	}))

	var originalGoPackageOpt = ""
	var genClient = false
	proto.Walk(parsed, proto.WithOption(func(i *proto.Option) {
		if i.Name == "go_package" {
			originalGoPackageOpt = i.Constant.Source
		} else if i.Name == "(gen_client)" {
			genClient = i.Constant.Source == "true"
		}
	}))

	parsedPath := strings.Split(path, "/")
	currentDir := root
	for _, element := range parsedPath {
		if element == "" {
			continue
		}
		if strings.HasSuffix(element, protoExtension) {
			// it's the file... just adding it and that's it
			currentDir.Files = append(currentDir.Files, ProtoFile{
				FileName:        ospathlib.GetFileName(path),
				FileDirPath:     ospathlib.GetFileDirPath(path),
				FileOSPath:      path,
				Imports:         imports,
				Services:        services,
				GenerateClient:  genClient,
				SourceGoPackage: originalGoPackageOpt,
				ProtoPackage:    protoPackage,
			})
			break
		}
		if currentDir.Name == element {
			continue
		}
		// looking if subDir was created
		found := false
		for _, subDir := range currentDir.SubDirs {
			if subDir.Name == element {
				currentDir = subDir
				found = true
				break
			}
		}

		if !found {
			// adding current element to our subtree
			newNode := &DirNode{
				Name:       element,
				DirAbsPath: fmt.Sprintf("%s/%s", currentDir.DirAbsPath, element),
			}
			currentDir.SubDirs = append(currentDir.SubDirs, newNode)
			currentDir = newNode
		}
	}

	return nil
}
