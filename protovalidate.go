package main

import (
	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/getkin/kin-openapi/openapi3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func getConstraints(field protoreflect.FieldDescriptor, schema *openapi3.Schema) {
	exprs, _ := proto.GetExtension(field.Options(), validate.E_Field).(*validate.FieldConstraints)
	if exprs == nil {
		return
	}

	switch v := exprs.Type.(type) {
	case *validate.FieldConstraints_Float:
		if c := float64(v.Float.GetConst()); c != 0 {
			schema.Min = &c
			schema.Max = &c
		}

		if lt := float64(v.Float.GetLt()); lt != 0 {
			schema.ExclusiveMax = true
			schema.Max = &lt
		} else if lte := float64(v.Float.GetLte()); lte != 0 {
			schema.Max = &lte
		}

		if gt := float64(v.Float.GetGt()); gt != 0 {
			schema.ExclusiveMin = true
			schema.Min = &gt
		} else if gte := float64(v.Float.GetGte()); gte != 0 {
			schema.Min = &gte
		}
	case *validate.FieldConstraints_Double:
		if c := float64(v.Double.GetConst()); c != 0 {
			schema.Min = &c
			schema.Max = &c
		}

		if lt := float64(v.Double.GetLt()); lt != 0 {
			schema.ExclusiveMax = true
			schema.Max = &lt
		} else if lte := float64(v.Double.GetLte()); lte != 0 {
			schema.Max = &lte
		}

		if gt := float64(v.Double.GetGt()); gt != 0 {
			schema.ExclusiveMin = true
			schema.Min = &gt
		} else if gte := float64(v.Double.GetGte()); gte != 0 {
			schema.Min = &gte
		}
	case *validate.FieldConstraints_Int32:
		if c := float64(v.Int32.GetConst()); c != 0 {
			schema.Min = &c
			schema.Max = &c
		}

		if lt := float64(v.Int32.GetLt()); lt != 0 {
			schema.ExclusiveMax = true
			schema.Max = &lt
		} else if lte := float64(v.Int32.GetLte()); lte != 0 {
			schema.Max = &lte
		}

		if gt := float64(v.Int32.GetGt()); gt != 0 {
			schema.ExclusiveMin = true
			schema.Min = &gt
		} else if gte := float64(v.Int32.GetGte()); gte != 0 {
			schema.Min = &gte
		}
	case *validate.FieldConstraints_Int64:
		if c := float64(v.Int64.GetConst()); c != 0 {
			schema.Min = &c
			schema.Max = &c
		}

		if lt := float64(v.Int64.GetLt()); lt != 0 {
			schema.ExclusiveMax = true
			schema.Max = &lt
		} else if lte := float64(v.Int64.GetLte()); lte != 0 {
			schema.Max = &lte
		}

		if gt := float64(v.Int64.GetGt()); gt != 0 {
			schema.ExclusiveMin = true
			schema.Min = &gt
		} else if gte := float64(v.Int64.GetGte()); gte != 0 {
			schema.Min = &gte
		}
	case *validate.FieldConstraints_Uint32:
		if c := float64(v.Uint32.GetConst()); c > 0 {
			schema.Min = &c
			schema.Max = &c
		}

		if lt := float64(v.Uint32.GetLt()); lt > 0 {
			schema.ExclusiveMax = true
			schema.Max = &lt
		} else if lte := float64(v.Uint32.GetLte()); lte > 0 {
			schema.Max = &lte
		}

		if gt := float64(v.Uint32.GetGt()); gt > 0 {
			schema.ExclusiveMin = true
			schema.Min = &gt
		} else if gte := float64(v.Uint32.GetGte()); gte > 0 {
			schema.Min = &gte
		}
	case *validate.FieldConstraints_Uint64:
		if c := float64(v.Uint64.GetConst()); c > 0 {
			schema.Min = &c
			schema.Max = &c
		}

		if lt := float64(v.Uint64.GetLt()); lt > 0 {
			schema.ExclusiveMax = true
			schema.Max = &lt
		} else if lte := float64(v.Uint64.GetLte()); lte > 0 {
			schema.Max = &lte
		}

		if gt := float64(v.Uint64.GetGt()); gt > 0 {
			schema.ExclusiveMin = true
			schema.Min = &gt
		} else if gte := float64(v.Uint64.GetGte()); gte > 0 {
			schema.Min = &gte
		}
	case *validate.FieldConstraints_Repeated:
		schema.UniqueItems = v.Repeated.GetUnique()
		schema.MinItems = v.Repeated.GetMinItems()
		schema.MaxItems = v.Repeated.MaxItems
		// TODO
	case *validate.FieldConstraints_Map:
		schema.MinProps = v.Map.GetMinPairs()
		schema.MaxProps = v.Map.MaxPairs
	case *validate.FieldConstraints_String_:
		if v.String_.Len != nil {
			schema.MinLength = v.String_.GetLen()
			schema.MaxLength = v.String_.Len
		} else {
			schema.MinLength = v.String_.GetMinLen()
			schema.MaxLength = v.String_.MaxLen
		}
		schema.Pattern = v.String_.GetPattern()

		switch {
		case v.String_.GetEmail():
			schema.Format = "email"
		case v.String_.GetHostname():
			schema.Format = "hostname"
		case v.String_.GetIpv4():
			schema.Format = "ipv4"
		case v.String_.GetIpv6():
			schema.Format = "ipv6"
		case v.String_.GetUri():
			schema.Format = "uri"
		case v.String_.GetUuid():
			schema.Format = "uuid"
		}
	}
}
