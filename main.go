package main

import (
	"flag"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

type param struct {
	Template *string
}

func main() {
	var flags flag.FlagSet
	param := param{
		Template: flags.String(
			"template", "./protoc-gen-connect-openapi_template.yaml", "Template file path"),
	}
	opts := &protogen.Options{
		ParamFunc: flags.Set,
	}

	opts.Run(func(plugin *protogen.Plugin) error {
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		for _, f := range plugin.Files {
			if !f.Generate {
				continue
			}

			if err := generateFile(plugin, f, param); err != nil {
				return err
			}
		}

		return nil
	})
}
