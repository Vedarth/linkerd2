package config

import (
	"io/ioutil"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	pb "github.com/linkerd/linkerd2/controller/gen/config"
	log "github.com/sirupsen/logrus"
)

// Global returns the Global protobuf config from the linkerd-config ConfigMap
func Global(filepath string) (*pb.Global, error) {
	config := &pb.Global{}
	err := unmarshalFile(filepath, config)
	return config, err
}

// Proxy returns the Proxy protobuf config from the linkerd-config ConfigMap
func Proxy(filepath string) (*pb.Proxy, error) {
	config := &pb.Proxy{}
	err := unmarshalFile(filepath, config)
	return config, err
}

func unmarshalFile(filepath string, msg proto.Message) error {
	configJSON, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	log.Debugf("%s config JSON: %s", filepath, configJSON)
	return unmarshal(string(configJSON), msg)
}

func unmarshal(json string, msg proto.Message) error {
	u := jsonpb.Unmarshaler{AllowUnknownFields: true}
	return u.Unmarshal(strings.NewReader(json), msg)
}

// FromConfigMap builds a configuration by reading a map with the keys "global"
// and "proxy", each containing JSON values.
func FromConfigMap(configMap map[string]string) (*pb.All, error) {
	c := &pb.All{Global: &pb.Global{}, Proxy: &pb.Proxy{}, Install: &pb.Install{}}

	if err := unmarshal(configMap["global"], c.Global); err != nil {
		return nil, err
	}

	if err := unmarshal(configMap["proxy"], c.Proxy); err != nil {
		return nil, err
	}

	if err := unmarshal(configMap["install"], c.Install); err != nil {
		return nil, err
	}

	return c, nil
}

// ToJSON encode the configuration to JSON, i.e. to be stored in a ConfigMap.
func ToJSON(configs *pb.All) (global, proxy, install string, err error) {
	m := jsonpb.Marshaler{EmitDefaults: true}

	global, err = m.MarshalToString(configs.GetGlobal())
	if err != nil {
		return
	}

	proxy, err = m.MarshalToString(configs.GetProxy())
	if err != nil {
		return
	}

	install, err = m.MarshalToString(configs.GetInstall())
	return
}

func unmarshal(json string, msg proto.Message) error {
	u := jsonpb.Unmarshaler{AllowUnknownFields: true}
	return u.Unmarshal(strings.NewReader(json), msg)
}

// FromConfigMap builds a configuration by reading a map with the keys "global"
// and "proxy", each containing JSON values.
func FromConfigMap(configMap map[string]string) (*pb.All, error) {
	c := &pb.All{Global: &pb.Global{}, Proxy: &pb.Proxy{}, Install: &pb.Install{}}

	if err := unmarshal(configMap["global"], c.Global); err != nil {
		return nil, err
	}

	if err := unmarshal(configMap["proxy"], c.Proxy); err != nil {
		return nil, err
	}

	if err := unmarshal(configMap["install"], c.Install); err != nil {
		return nil, err
	}

	return c, nil
}

// ToJSON encode the configuration to JSON, i.e. to be stored in a ConfigMap.
func ToJSON(configs *pb.All) (global, proxy, install string, err error) {
	m := jsonpb.Marshaler{EmitDefaults: true}

	global, err = m.MarshalToString(configs.GetGlobal())
	if err != nil {
		return
	}

	proxy, err = m.MarshalToString(configs.GetProxy())
	if err != nil {
		return
	}

	install, err = m.MarshalToString(configs.GetInstall())
	return
}
