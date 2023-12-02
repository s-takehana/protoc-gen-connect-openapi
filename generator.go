package main

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/invopop/yaml"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func generateFile(plugin *protogen.Plugin, file *protogen.File, param param) error {
	loader := openapi3.NewLoader()
	t, err := loader.LoadFromFile(*param.Template)
	if err != nil {
		return err
	}

	schemas, err := generateComponentsSchemas(file)
	if err != nil {
		return err
	}

	paths, tags, err := generatePathsAndTags(file)
	if err != nil {
		return err
	}

	t.OpenAPI = "3.0.3"
	t.Components.Schemas = *schemas
	t.Paths = paths
	t.Tags = tags

	bytes, err := yaml.Marshal(&t)
	if err != nil {
		return err
	}

	filename := strings.TrimSuffix(
		strings.ReplaceAll(file.Desc.Path(), "/", "."), filepath.Ext(file.Desc.Path())) + ".pb.yaml"

	g := plugin.NewGeneratedFile(filename, "")
	if _, err := g.Write(bytes); err != nil {
		return err
	}

	return nil
}

func generateComponentsSchemas(file *protogen.File) (*openapi3.Schemas, error) {
	schemas := make(openapi3.Schemas, len(file.Messages)+1)
	for _, m := range file.Messages {

		properties, err := generateProperties(m)
		if err != nil {
			return nil, err
		}

		k := string(m.Desc.FullName())
		schemas[k] = &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:        openapi3.TypeObject,
				Description: generateCommentsString(m.Comments),
				Properties:  *properties,
			},
		}

	}
	schemas["DefaultErrorResponse"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: openapi3.TypeObject,
			Properties: openapi3.Schemas{
				"code": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: openapi3.TypeString,
						Enum: []any{
							"canceled", "unknown", "invalid_argument", "deadline_exceeded", "not_found",
							"already_exists", "permission_denied", "resource_exhausted", "failed_precondition",
							"aborted", "out_of_range", "unimplemented", "internal", "unavailable", "data_loss",
							"unauthenticated",
						},
					},
				},
				"message": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: openapi3.TypeString,
					},
				},
			},
		},
	}
	return &schemas, nil
}

func generateProperties(message *protogen.Message) (*openapi3.Schemas, error) {
	fields := make(openapi3.Schemas, len(message.Fields))
	for _, f := range message.Fields {

		switch {
		case f.Desc.Kind() == protoreflect.EnumKind:

			enum := make([]any, len(f.Enum.Values))
			for i, e := range f.Enum.Values {
				enum[i] = e.Desc.Name()
			}

			fields[string(f.Desc.Name())] = &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:        openapi3.TypeString,
					Description: generateCommentsString(f.Enum.Comments),
					Enum:        enum,
				},
			}

		case f.Desc.IsMap():

			sch := &openapi3.Schema{
				Type:        openapi3.TypeObject,
				Description: generateCommentsString(f.Comments),
			}

			getConstraints(f.Desc, sch)

			switch {
			case f.Desc.MapKey().Kind() != protoreflect.StringKind:
				has := true
				sch.AdditionalProperties = openapi3.AdditionalProperties{
					Has: &has,
				}

				fields[string(f.Desc.Name())] = &openapi3.SchemaRef{
					Value: sch,
				}
			case f.Desc.MapValue().Kind() == protoreflect.EnumKind,
				f.Desc.MapValue().Kind() == protoreflect.MessageKind:
				// TODO
			default:
				t, format, err := convertToTypeAndFormat(f.Desc.MapValue().Kind())
				if err != nil {
					return nil, err
				}

				sch.AdditionalProperties = openapi3.AdditionalProperties{
					Schema: &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:   t,
							Format: format,
						},
					},
				}

				fields[string(f.Desc.Name())] = &openapi3.SchemaRef{
					Value: sch,
				}
			}

		case f.Desc.Kind() == protoreflect.MessageKind:

			switch string(f.Message.Desc.FullName()) {
			case "google.protobuf.Duration":
				fields[string(f.Desc.Name())] = &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:        openapi3.TypeString,
						Description: generateCommentsString(f.Comments),
					},
				}
			case "google.protobuf.Timestamp":
				fields[string(f.Desc.Name())] = &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:        openapi3.TypeString,
						Format:      "date-time",
						Description: generateCommentsString(f.Comments),
					},
				}
			case string(f.Parent.Desc.FullName()):
				t := openapi3.TypeObject
				if f.Desc.IsList() {
					t = openapi3.TypeArray
				}

				fields[string(f.Desc.Name())] = &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:        t,
						Description: generateCommentsString(f.Comments),
					},
				}
			default:
				properties, err := generateProperties(f.Message)
				if err != nil {
					return nil, err
				}

				if f.Desc.IsList() {
					fields[string(f.Desc.Name())] = &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:        openapi3.TypeArray,
							Description: generateCommentsString(f.Comments),
							Items: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type:       openapi3.TypeObject,
									Properties: *properties,
								},
							},
						},
					}
				} else {
					fields[string(f.Desc.Name())] = &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:        openapi3.TypeObject,
							Description: generateCommentsString(f.Comments),
							Properties:  *properties,
						},
					}
				}
			}

		default:

			t, format, err := convertToTypeAndFormat(f.Desc.Kind())
			if err != nil {
				return nil, err
			}

			var schema *openapi3.Schema
			if f.Desc.IsList() {
				schema = &openapi3.Schema{
					Type:        openapi3.TypeArray,
					Description: generateCommentsString(f.Comments),
					Items: &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:   t,
							Format: format,
						},
					},
				}
			} else {
				schema = &openapi3.Schema{
					Type:        t,
					Format:      format,
					Description: generateCommentsString(f.Comments),
				}
			}

			getConstraints(f.Desc, schema)

			fields[string(f.Desc.Name())] = &openapi3.SchemaRef{
				Value: schema,
			}

		}

	}
	return &fields, nil
}

func generatePathsAndTags(file *protogen.File) (*openapi3.Paths, openapi3.Tags, error) {
	var paths openapi3.Paths
	tags := make([]*openapi3.Tag, len(file.Services))
	for i, s := range file.Services {
		for _, m := range s.Methods {

			ret, err := url.JoinPath("/", string(s.Desc.FullName()), string(m.Desc.Name()))
			if err != nil {
				return nil, nil, err
			}

			resDesc := generateCommentsString(m.Output.Comments)
			defaultResDesc := "Default error response."

			var responses openapi3.Responses
			responses.Set("200", &openapi3.ResponseRef{
				Value: &openapi3.Response{
					Description: &resDesc,
					Content: map[string]*openapi3.MediaType{
						"application/json": {
							Schema: &openapi3.SchemaRef{
								Ref: "#/components/schemas/" + string(m.Output.Desc.FullName()),
							},
						},
					},
				},
			})
			responses.Set("default", &openapi3.ResponseRef{
				Value: &openapi3.Response{
					Description: &defaultResDesc,
					Content: map[string]*openapi3.MediaType{
						"application/json": {
							Schema: &openapi3.SchemaRef{
								Ref: "#/components/schemas/DefaultErrorResponse",
							},
						},
					},
				},
			})

			paths.Set(ret, &openapi3.PathItem{
				Post: &openapi3.Operation{
					Tags:        []string{string(s.Desc.FullName())},
					Description: generateCommentsString(m.Comments),
					OperationID: strings.Join([]string{string(file.GoPackageName), s.GoName, m.GoName}, "_"),
					RequestBody: &openapi3.RequestBodyRef{
						Value: &openapi3.RequestBody{
							Content: map[string]*openapi3.MediaType{
								"application/json": {
									Schema: &openapi3.SchemaRef{
										Ref: "#/components/schemas/" + string(m.Input.Desc.FullName()),
									},
								},
							},
						},
					},
					Responses: &responses,
				},
			})

		}
		tags[i] = &openapi3.Tag{
			Name: string(s.Desc.FullName()),
		}
	}
	return &paths, tags, nil
}

func generateCommentsString(comments protogen.CommentSet) string {
	l := comments.Leading.String()
	t := comments.Trailing.String()
	if len(l) == 0 && len(t) == 0 {
		return ""
	}

	r := strings.NewReplacer("// ", "", "//", "")
	return r.Replace(l) + r.Replace(t)
}

func convertToTypeAndFormat(kind protoreflect.Kind) (string, string, error) {
	switch kind {
	case protoreflect.DoubleKind:
		return openapi3.TypeNumber, "double", nil
	case protoreflect.FloatKind:
		return openapi3.TypeNumber, "float", nil
	case protoreflect.Int32Kind, protoreflect.Uint32Kind, protoreflect.Sint32Kind,
		protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind:
		return openapi3.TypeInteger, "int32", nil
	case protoreflect.Int64Kind, protoreflect.Uint64Kind, protoreflect.Sint64Kind,
		protoreflect.Fixed64Kind, protoreflect.Sfixed64Kind:
		return openapi3.TypeInteger, "int64", nil
	case protoreflect.BoolKind:
		return openapi3.TypeBoolean, "", nil
	case protoreflect.StringKind:
		return openapi3.TypeString, "", nil
	case protoreflect.BytesKind:
		return openapi3.TypeString, "byte", nil
	case protoreflect.GroupKind:
		return "", "", errors.New("groups are a deprecated feature that should not be used")
	case protoreflect.EnumKind:
	case protoreflect.MessageKind:
	}

	return "", "", fmt.Errorf("can not convert kind: %d", kind)
}
