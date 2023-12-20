package types

import (
	"fmt"
	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type RdkDescriptor struct {
	DescriptorSource  grpcurl.DescriptorSource
	ServiceDescriptor desc.Descriptor
	Methods           protoreflect.MethodDescriptors
}

var Rdk *RdkDescriptor

func (r *RdkDescriptor) MethodList() []protoreflect.MethodDescriptor {
	var methods []protoreflect.MethodDescriptor
	for i := 0; i < r.Methods.Len(); i++ {
		methods = append(methods, r.Methods.Get(i))
	}
	return methods
}

func (r *RdkDescriptor) MethodNames() []string {
	var names []string
	for i := 0; i < r.Methods.Len(); i++ {
		names = append(names, string(r.Methods.Get(i).Name()))
	}
	return names
}

func parseProtobufDescriptor() error {
	sets, err := grpcurl.DescriptorSourceFromProtoSets("protobuf.descriptor.bin")
	if err != nil {
		return err
	}

	services, err := sets.ListServices()
	if err != nil {
		return err
	}
	var kyverdk string
	for _, service := range services {
		if service == "kyverdk.runtime.v1.RuntimeService" {
			kyverdk = service
		}
	}
	if kyverdk == "" {
		return fmt.Errorf("kyverdk.runtime.v1.RuntimeService not found")
	}

	serviceDescriptor, err := sets.FindSymbol(kyverdk)
	if err != nil {
		return err
	}

	Rdk = &RdkDescriptor{
		DescriptorSource:  sets,
		ServiceDescriptor: serviceDescriptor,
		Methods:           serviceDescriptor.GetFile().GetServices()[0].UnwrapService().Methods(),
	}

	return nil
}

func init() {
	err := parseProtobufDescriptor()
	if err != nil {
		panic(err)
	}
}
