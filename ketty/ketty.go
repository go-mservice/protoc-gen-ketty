// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2015 The Go Authors.  All rights reserved.
// https://github.com/golang/protobuf
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//     * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package ketty outputs gRPC service descriptions in Go code.
// It runs as a plugin for the Go protocol buffer compiler plugin.
// It is linked in to protoc-gen-go.
package ketty

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"github.com/yyzybb537/ketty/log"
	"github.com/golang/protobuf/proto"

	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	kettyProto "github.com/yyzybb537/protoc-gen-ketty/include"
)

var _ = log.GetLog

// generatedCodeVersion indicates a version of the generated code.
// It is incremented whenever an incompatibility between the generated code and
// the ketty package is introduced; the generated code references
// a constant, ketty.SupportPackageIsVersionN (where N is generatedCodeVersion).
const generatedCodeVersion = 4

// Paths for packages used by code generated in this file,
// relative to the import_prefix of the generator.Generator.
const (
	kettyPkgPath    = "github.com/yyzybb537/ketty"
)

func init() {
	generator.RegisterPlugin(new(ketty))
}

// ketty is an implementation of the Go protocol buffer compiler's
// plugin architecture.  It generates bindings for gRPC support.
type ketty struct {
	gen *generator.Generator
}

// Name returns the name of this plugin, "ketty".
func (g *ketty) Name() string {
	return "ketty"
}

// The names for packages imported in the generated code.
// They may vary from the final path component of the import path
// if the name is used by other packages.
var (
	contextPkg string
	kettyPkg    string
)

// Init initializes the plugin.
func (g *ketty) Init(gen *generator.Generator) {
	g.gen = gen
	contextPkg = "context"
	kettyPkg = generator.RegisterUniquePackageName("ketty", nil)
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (g *ketty) objectNamed(name string) generator.Object {
	g.gen.RecordTypeUse(name)
	return g.gen.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (g *ketty) typeName(str string) string {
	return g.gen.TypeName(g.objectNamed(str))
}

// P forwards to g.gen.P.
func (g *ketty) P(args ...interface{}) { g.gen.P(args...) }

// Generate generates code for the services in the given file.
func (g *ketty) Generate(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	println("ketty.Generate")

	g.P("// Reference imports to suppress errors if they are not otherwise used.")
	g.P("var _ ", kettyPkg, ".Dummy")
	g.P()

	// Assert version compatibility.
	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the ketty package it is being compiled against.")
	//g.P("const _ = ", kettyPkg, ".SupportPackageIsVersion", generatedCodeVersion)
	g.P()

	for i, service := range file.FileDescriptorProto.Service {
		g.generateService(file, service, i)
	}

	for _, message := range file.FileDescriptorProto.MessageType {
		//println("------ message ", *message.Name)
		//println(log.LogFormat(message, log.Indent))
		if message.Options == nil {
			continue
        }

		opts := getKettyOptions(message)
		//println(log.LogFormat(opts, log.Indent))

		g.generateOptionMethods(message, opts)
    }
}

type kettyOptions struct {
	isUseKettyHttpExtend bool
	transport string
	marshal string
}

func getKettyOptions(message *pb.DescriptorProto) (opts *kettyOptions) {
	opts = &kettyOptions{}
	iisUseKettyHttpExtend, err := proto.GetExtension(message.Options, kettyProto.E_UseKettyHttpExtend)
	if err == nil {
		if iisUseKettyHttpExtend.(*bool) != nil {
			opts.isUseKettyHttpExtend = *iisUseKettyHttpExtend.(*bool)
		}
	}

	iTransport, err := proto.GetExtension(message.Options, kettyProto.E_Transport)
	if err == nil {
		if iTransport.(*string) != nil {
			opts.transport = *iTransport.(*string)
        }
	}

	iMarshal, err := proto.GetExtension(message.Options, kettyProto.E_Marshal)
	if err == nil {
		if iMarshal.(*string) != nil {
			opts.marshal = *iMarshal.(*string)
        }
	}

	return
}

func (g *ketty) generateOptionMethods(message *pb.DescriptorProto, opts *kettyOptions) {
	if opts.isUseKettyHttpExtend {
		g.P("func (*", message.Name, ") KettyHttpExtendMessage() {}")
		g.P()
    }

	if opts.marshal != "" {
		g.P("func (*", message.Name, ") KettyMarshal() string {")
		g.P("return \"", opts.marshal, "\"")
		g.P("}")
		g.P()
    }

	if opts.transport != "" {
		g.P("func (*", message.Name, ") KettyTransport() string {")
		g.P("return \"", opts.transport, "\"")
		g.P("}")
		g.P()
    }
}

// GenerateImports generates the import declaration for this file.
func (g *ketty) GenerateImports(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	g.P("import (")
	g.P(kettyPkg, " ", strconv.Quote(path.Join(g.gen.ImportPrefix, kettyPkgPath)))
	g.P(")")
	g.P()
}

// reservedClientName records whether a client name is reserved on the client side.
var reservedClientName = map[string]bool{
// TODO: do we need any in gRPC?
}

func unexport(s string) string { return strings.ToLower(s[:1]) + s[1:] }

// generateService generates all the code for the named service.
func (g *ketty) generateService(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto, index int) {
	//path := fmt.Sprintf("6,%d", index) // 6 means service.

	origServName := service.GetName()
	fullServName := origServName
	if pkg := file.GetPackage(); pkg != "" {
		fullServName = pkg + "." + fullServName
	}
	servName := generator.CamelCase(origServName)

	// ketty handle
	handleT := servName + "HandleT"
	g.P(fmt.Sprintf("type %s struct {", handleT))
	g.P("desc *grpc.ServiceDesc")
	g.P("}")
	g.P()

	g.P(fmt.Sprintf("func (h *%s) Implement() interface{} {", handleT))
	g.P("return h.desc")
	g.P("}")
	g.P()

	g.P(fmt.Sprintf("func (h *%s) ServiceName() string {", handleT))
	g.P("return h.desc.ServiceName")
	g.P("}")
	g.P()

	g.P(fmt.Sprintf("var %sHandle = &%s{ desc : &_%s_serviceDesc }", servName, handleT, servName))
	g.P()

	// ketty Client struct.
	g.P("type Ketty", servName, "Client struct {")
	g.P("client ", kettyPkg, ".Client")
	g.P("}")
	g.P()

	// NewClient factory.
	g.P("func NewKetty", servName, "Client (client ", kettyPkg, ".Client) ", "*Ketty", servName, "Client {")
	g.P("return &Ketty", servName, "Client{client}")
	g.P("}")
	g.P()

	// Methods
	var methodIndex, streamIndex int
	serviceDescVar := "_" + servName + "_serviceDesc"
	
	for _, method := range service.Method {
		var descExpr string
		if !method.GetServerStreaming() && !method.GetClientStreaming() {
			// Unary RPC method
			descExpr = fmt.Sprintf("&%s.Methods[%d]", serviceDescVar, methodIndex)
			methodIndex++
		} else {
			// Streaming RPC method
			descExpr = fmt.Sprintf("&%s.Streams[%d]", serviceDescVar, streamIndex)
			streamIndex++
		}
		g.generateClientMethod(servName, fullServName, serviceDescVar, method, descExpr)
	}
}

// generateClientSignature returns the client-side signature for a method.
func (g *ketty) generateClientSignature(servName string, method *pb.MethodDescriptorProto) string {
	origMethName := method.GetName()
	methName := generator.CamelCase(origMethName)
	if reservedClientName[methName] {
		methName += "_"
	}
	reqArg := ", in *" + g.typeName(method.GetInputType())
	if method.GetClientStreaming() {
		reqArg = ""
	}
	respName := "*" + g.typeName(method.GetOutputType())
	if method.GetServerStreaming() || method.GetClientStreaming() {
		respName = servName + "_" + generator.CamelCase(origMethName) + "Client"
	}
	return fmt.Sprintf("%s(ctx %s.Context%s) (%s, error)", methName, contextPkg, reqArg, respName)
}

func (g *ketty) generateClientMethod(servName, fullServName, serviceDescVar string, method *pb.MethodDescriptorProto, descExpr string) {
	//sname := fmt.Sprintf("/%s/%s", fullServName, method.GetName())
	methName := generator.CamelCase(method.GetName())
	//inType := g.typeName(method.GetInputType())
	outType := g.typeName(method.GetOutputType())

	g.P("func (this *Ketty", servName, "Client) ", g.generateClientSignature(servName, method), "{")
	g.P("out := new(", outType, ")")
	g.P("err := ", fmt.Sprintf("this.client.Invoke(ctx, %sHandle, \"%s\", in, out)", servName, methName))
	g.P("if err != nil { return nil, err }")
	g.P("return out, nil")
	g.P("}")
	g.P()
	return
}


